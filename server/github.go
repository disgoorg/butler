package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/webhook"
	"github.com/google/go-github/v93/github"

	"github.com/disgoorg/disgo-butler/server/database"
)

var (
	markdownHeaderRegex            = regexp.MustCompile(`[ \t]*#+[ \t]+([^\r\n]+)`)
	markdownBulletRegex            = regexp.MustCompile(`([ \t]*)[*|-][ \t]+([^\r\n]+)`)
	markdownCheckBoxCheckedRegex   = regexp.MustCompile(`([ \t]*)[*|-][ \t]{0,4}\[x][ \t]+([^\r\n]+)`)
	markdownCheckBoxUncheckedRegex = regexp.MustCompile(`([ \t]*)[*|-][ \t]{0,4}\[ ][ \t]+([^\r\n]+)`)
	prURLRegex                     = regexp.MustCompile(`https?://github\.com/(\w+/\w+)/pull/(\d+)`)
	commitURLRegex                 = regexp.MustCompile(`https?://github\.com/\w+/\w+/commit/([a-f\d]{7})[a-f\d]+`)
	mentionRegex                   = regexp.MustCompile(`@(\w+)`)
)

func (s *Server) handleGithubWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(s.Cfg.GithubWebhookSecret))
	if err != nil {
		slog.Error("failed to validate payload", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		slog.Error("failed to parse webhook", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch release := event.(type) {
	case *github.ReleaseEvent:
		err = s.processReleaseEvent(release)
	}
	if err != nil {
		slog.Error("failed to process webhook", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) processReleaseEvent(event *github.ReleaseEvent) error {
	if event.GetAction() != "published" {
		return nil
	}

	org := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	fullName := event.GetRepo().GetFullName()

	cfg, err := s.DB.GetGithubRelease(context.Background(), fullName)
	if database.IsNotFound(err) {
		return errors.New("no config found for this repo")
	} else if err != nil {
		return err
	}

	s.webhooksMu.Lock()
	webhookClient, ok := s.Webhooks[fullName]
	if !ok {
		webhookClient = webhook.New(cfg.WebhookID, cfg.WebhookToken)
		s.Webhooks[fullName] = webhookClient
	}
	s.webhooksMu.Unlock()

	releases, _, err := s.GitHub.Repositories.ListReleases(context.Background(), org, repo, &github.ListOptions{PerPage: 2})
	if err != nil {
		return err
	}
	var previousRelease *github.RepositoryRelease
	for _, release := range releases {
		if release.GetTagName() != event.GetRelease().GetTagName() {
			previousRelease = release
			break
		}
	}
	if previousRelease == nil {
		return errors.New("no previous release found")
	}

	comparison, _, err := s.GitHub.Repositories.CompareCommits(context.Background(), org, repo, previousRelease.GetTagName(), event.GetRelease().GetTagName(), nil)
	if err != nil {
		return err
	}

	message := parseMarkdown(event.GetRelease().GetBody())
	if len(message) > 1024 {
		message = substr(message, 0, 1024)
		if index := strings.LastIndex(message, "\n"); index != -1 {
			message = message[:index]
		}
		message += "\n…"
	}
	message += "\n\n__**Commits:**__\n"
	for _, commit := range comparison.Commits {
		commitLine := strings.SplitN(commit.GetCommit().GetMessage(), "\n", 2)[0]
		shortID := substr(commit.GetSHA(), 0, 7)
		line := fmt.Sprintf("[`%s`](%s) %s\n", shortID, commit.GetHTMLURL(), commitLine)
		if len(message)+len(line) > 4068 {
			message += "…"
			break
		}
		message += line
	}

	createdAt := event.GetRelease().GetCreatedAt().Time
	msg, err := webhookClient.CreateMessage(discord.NewWebhookMessageCreate().
		WithContent(discord.RoleMention(cfg.PingRole)).
		WithEmbeds(discord.NewEmbed().
			WithAuthor(
				fmt.Sprintf("%s version %s has been released", repo, event.GetRelease().GetTagName()),
				event.GetRelease().GetHTMLURL(),
				event.GetRepo().GetOwner().GetAvatarURL(),
			).
			WithDescription(message).
			WithColor(0x5865f2).
			WithFooter("Release by "+event.GetRelease().GetAuthor().GetLogin(), event.GetRelease().GetAuthor().GetAvatarURL()).
			WithTimestamp(createdAt),
		),
		rest.CreateWebhookMessageParams{Wait: true},
	)
	if err != nil {
		return err
	}
	_, err = s.Client.Rest.CrosspostMessage(msg.ChannelID, msg.ID)
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
	text = prURLRegex.ReplaceAllString(text, "[$1#$2]($0)")
	text = commitURLRegex.ReplaceAllString(text, "[`$1`]($0)")
	text = mentionRegex.ReplaceAllString(text, "[@$1](https://github.com/$1)")
	return text
}
