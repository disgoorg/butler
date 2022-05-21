package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/google/go-github/v44/github"
)

var (
	markdownHeaderRegex            = regexp.MustCompile(`[ \t]*#+[ \t]+([^\r\n]+)`)
	markdownBulletRegex            = regexp.MustCompile(`([ \t]*)[*|-][ \t]+([^\r\n]+)`)
	markdownCheckBoxCheckedRegex   = regexp.MustCompile(`([ \t]*)[*|-][ \t]{0,4}\[x][ \t]+([^\r\n]+)`)
	markdownCheckBoxUncheckedRegex = regexp.MustCompile(`([ \t]*)[*|-][ \t]{0,4}\[ ][ \t]+([^\r\n]+)`)
)

func HandleGithubWebhook(b *butler.Butler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := github.ValidatePayload(r, []byte(b.Config.GithubWebhookSecret))
		if err != nil {
			b.Logger.Errorf("Failed to validate payload: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		event, err := github.ParseWebHook(github.WebHookType(r), payload)
		if err != nil {
			b.Logger.Errorf("Failed to parse webhook: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		switch e := event.(type) {
		case *github.ReleaseEvent:
			err = processReleaseEvent(b, e)
		}
		if err != nil {
			b.Logger.Errorf("Failed to process webhook: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func processReleaseEvent(b *butler.Butler, e *github.ReleaseEvent) error {
	if e.GetAction() != "published" {
		return nil
	}

	org, repo, fullName := e.GetRepo().GetOwner().GetLogin(), e.GetRepo().GetName(), e.GetRepo().GetFullName()

	cfg, ok := b.Config.GithubReleases[fullName]
	if !ok {
		return errors.New("no config found for this repo")
	}

	webhookClient, ok := b.Webhooks[fullName]
	if !ok {
		webhookClient = webhook.NewClient(cfg.WebhookID, cfg.WebhookToken)
		b.Webhooks[fullName] = webhookClient
	}

	releases, _, err := b.GitHubClient.Repositories.ListReleases(context.TODO(), org, repo, &github.ListOptions{PerPage: 2})
	if err != nil {
		return err
	}
	var previousRelease *github.RepositoryRelease
	for _, release := range releases {
		if release.GetTagName() != e.GetRelease().GetTagName() {
			previousRelease = release
			break
		}
	}

	comparison, _, err := b.GitHubClient.Repositories.CompareCommits(context.TODO(), org, repo, previousRelease.GetTagName(), e.GetRelease().GetTagName(), nil)
	if err != nil {
		return err
	}

	message := parseMarkdown(e.GetRelease().GetBody())
	if len(message) > 1024 {
		message = substr(message, 0, 1024) + "…"
	}
	message += "\n\n__**Commits:**__\n"
out:
	for _, commit := range comparison.Commits {
		commitLines := strings.Split(commit.GetCommit().GetMessage(), "\n")
		for i, commitLine := range commitLines {
			shortId := "......."
			if i == 0 {
				shortId = substr(commit.GetSHA(), 0, 7)
			}
			line := fmt.Sprintf("[`%s`](%s) %s\n", shortId, commit.GetURL(), commitLine)
			if len(message)+len(line) > 4068 {
				message += "…"
				break out
			}
			message += line
		}
	}

	_, err = webhookClient.CreateMessage(discord.NewWebhookMessageCreateBuilder().
		SetContent(discord.RoleMention(cfg.PingRole)).
		SetEmbeds(discord.NewEmbedBuilder().
			SetAuthor(
				fmt.Sprintf("%s version %s has been released", repo, e.Release.GetTagName()),
				e.GetRelease().GetHTMLURL(),
				e.GetRepo().GetOwner().GetAvatarURL(),
			).
			SetDescription(message).
			SetColor(0x5865f2).
			SetFooter("Release by "+e.GetRelease().GetAuthor().GetLogin(), e.GetRelease().GetAuthor().GetAvatarURL()).
			SetTimestamp(e.GetRelease().GetCreatedAt().Time).
			Build(),
		).
		Build(),
	)
	return err
}

func substr(input string, start int, length int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if start+length > len(asRunes) {
		length = len(asRunes) - start
	}

	return string(asRunes[start : start+length])
}

func parseMarkdown(text string) string {
	text = markdownCheckBoxCheckedRegex.ReplaceAllString(text, "$1:ballot_box_with_check: $2")
	text = markdownCheckBoxUncheckedRegex.ReplaceAllString(text, "$1:white_square_button: $2")
	text = markdownHeaderRegex.ReplaceAllString(text, "**$1**")
	text = markdownBulletRegex.ReplaceAllString(text, "$1• $2")
	return text
}
