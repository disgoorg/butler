package routes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/disgoorg/disgo-butler/butler"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/google/go-github/v44/github"
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

func processReleaseEvent(b *butler.Butler, e *github.ReleaseEvent) error {
	if e.GetAction() != "created" {
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

	message := "__**Commits:**__\n"
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
