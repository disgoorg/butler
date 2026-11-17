package server

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/hhhapz/doc"
)

const (
	embedDescriptionFormat = "```go\n%s```\n\n%s\n%s"
	embedTitleFormat       = "%s: %s"
	embedColor             = 0x5865f2
	embedPackageURLFormat  = "https://pkg.go.dev/%s"
	embedURLFormat         = embedPackageURLFormat + "#%s"
	exampleFormat          = "\n%s:\n```go\n%s\n```\n```%s```\n"
	pkgInfo                = "<pkg_info>"
)

func getDocsEmbed(pkg doc.Package, query string, expandSignature bool, expandComment bool, expandMethods bool, expandExamples bool) (discord.Embed, discord.StringSelectMenuComponent) {
	var (
		embed         discord.Embed
		moreSignature bool
		moreMethods   bool
		moreComment   bool
		moreExamples  bool
	)

	if query == "" || query == pkgInfo {
		embed, moreComment, moreExamples = embedFromPackage(pkg, expandComment, expandExamples)
	} else {
		values := strings.Split(query, ".")
		for i := range values {
			values[i] = strings.ToLower(values[i])
		}
		if t, ok := pkg.Types[values[0]]; ok {
			if len(values) > 1 {
				if m, ok := t.Methods[values[1]]; ok {
					embed, moreSignature, moreComment = embedFromMethod(pkg, m, expandSignature, expandComment, expandExamples)
				}
			} else {
				embed, moreSignature, moreComment = embedFromType(pkg, t, expandSignature, expandComment, expandMethods, expandExamples)
				moreMethods = len(t.Methods) > 0 && !expandMethods
			}
		} else if f, ok := pkg.Functions[values[0]]; ok {
			embed, moreSignature, moreComment = embedFromFunc(pkg, f, expandSignature, expandComment, expandExamples)
		}
	}
	if len(embed.Description) > 4096 {
		embed.Description = embed.Description[:4095] + "…"
	}

	var options []discord.StringSelectMenuOption
	if moreSignature {
		options = append(options, discord.NewStringSelectMenuOption("expand signature", "expand:signature").WithEmoji(discord.ComponentEmoji{Name: "🔼"}))
	}
	if expandSignature {
		options = append(options, discord.NewStringSelectMenuOption("collapse signature", "collapse:signature").WithEmoji(discord.ComponentEmoji{Name: "🔽"}))
	}

	if moreMethods {
		options = append(options, discord.NewStringSelectMenuOption("expand methods", "expand:methods").WithEmoji(discord.ComponentEmoji{Name: "🔼"}))
	}
	if expandMethods {
		options = append(options, discord.NewStringSelectMenuOption("collapse methods", "collapse:methods").WithEmoji(discord.ComponentEmoji{Name: "🔽"}))
	}

	if moreComment {
		options = append(options, discord.NewStringSelectMenuOption("expand comment", "expand:comment").WithEmoji(discord.ComponentEmoji{Name: "🔼"}))
	}
	if expandComment {
		options = append(options, discord.NewStringSelectMenuOption("collapse comment", "collapse:comment").WithEmoji(discord.ComponentEmoji{Name: "🔽"}))
	}

	if moreExamples {
		options = append(options, discord.NewStringSelectMenuOption("expand examples", "expand:examples").WithEmoji(discord.ComponentEmoji{Name: "🔼"}))
	}
	if expandExamples {
		options = append(options, discord.NewStringSelectMenuOption("collapse examples", "collapse:examples").WithEmoji(discord.ComponentEmoji{Name: "🔽"}))
	}

	options = append(options, discord.NewStringSelectMenuOption("delete", "delete").WithEmoji(discord.ComponentEmoji{Name: "❌"}))

	return embed, discord.NewStringSelectMenu("/docs_action", "action", options...)
}

func embedFromPackage(pkg doc.Package, expandComment bool, expandExamples bool) (discord.Embed, bool, bool) {
	var (
		moreComment  bool
		moreExamples bool
	)

	description := pkg.Overview.Markdown()
	if !expandComment && len(description) > 1024 {
		description = description[:1023] + "…"
		moreComment = true
	}

	var examples string
	for _, example := range pkg.Examples {
		examples += fmt.Sprintf(exampleFormat, example.Name, example.Code, example.Output)
	}
	if !expandExamples && len(examples) > 1024 {
		examples = examples[:1023] + "…"
		moreExamples = true
	}

	return discord.Embed{
		Title:       pkg.URL,
		URL:         fmt.Sprintf(embedPackageURLFormat, pkg.URL),
		Description: description + "\n" + examples,
		Color:       embedColor,
	}, moreComment, moreExamples
}

func embedFromMethod(pkg doc.Package, method doc.Method, expandSignature bool, expandComment bool, expandExamples bool) (discord.Embed, bool, bool) {
	description, moreSignature, moreComment := formatDescription(method.Signature, method.Comment, method.Examples, expandSignature, expandComment, expandExamples)
	return discord.Embed{
		Title:       fmt.Sprintf(embedTitleFormat, pkg.URL, method.For+"."+method.Name),
		URL:         fmt.Sprintf(embedURLFormat, pkg.URL, method.For+"."+method.Name),
		Description: description,
		Color:       embedColor,
	}, moreSignature, moreComment
}

func embedFromFunc(pkg doc.Package, function doc.Function, expandSignature bool, expandComment bool, expandExamples bool) (discord.Embed, bool, bool) {
	description, moreSignature, moreComment := formatDescription(function.Signature, function.Comment, function.Examples, expandSignature, expandComment, expandExamples)
	return discord.Embed{
		Title:       fmt.Sprintf(embedTitleFormat, pkg.URL, function.Name),
		URL:         fmt.Sprintf(embedURLFormat, pkg.URL, function.Name),
		Description: description,
		Color:       embedColor,
	}, moreSignature, moreComment
}

func embedFromType(pkg doc.Package, t doc.Type, expandSignature bool, expandComment bool, expandMethods bool, expandExamples bool) (discord.Embed, bool, bool) {
	description, moreSignature, moreComment := formatDescription(t.Signature, t.Comment, t.Examples, expandSignature, expandComment, expandExamples)
	if expandMethods {
		methods := "```go\n"
		for _, method := range t.Methods {
			methods += method.Signature + "\n\n"
		}
		description += methods + "\n```"
		if len(description) > 4096 {
			description = description[:4082] + "…\n```"
		}
	}
	return discord.Embed{
		Title:       fmt.Sprintf(embedTitleFormat, pkg.URL, t.Name),
		URL:         fmt.Sprintf(embedURLFormat, pkg.URL, t.Name),
		Description: description,
		Color:       embedColor,
	}, moreSignature, moreComment
}

func formatDescription(signature string, comment doc.Comment, examples []doc.Example, expandSignature bool, expandComment bool, expandExamples bool) (string, bool, bool) {
	var (
		moreSignature bool
		moreComment   bool
	)
	lines := strings.Split(signature, "\n")
	if !expandSignature && len(lines) > 6 {
		moreSignature = true
		signature = lines[0] + "\n…"
	} else if len(signature) > 4096 {
		signature = signature[:4082] + "…\n```"
	}

	markdown := comment.Markdown()
	if markdown == "" {
		markdown = "No comments found."
	}
	lines = strings.Split(markdown, "\n")
	if !expandComment && len(lines) > 6 {
		moreComment = true
		markdown = lines[0] + "\n…"
	}
	if !expandComment && len(markdown) > 1024 {
		moreComment = true
		markdown = markdown[:1023] + "…"
	}

	var examplesStr string
	if expandExamples {
		for _, example := range examples {
			examplesStr += fmt.Sprintf(exampleFormat, example.Name, example.Code, example.Output)
		}
	}
	return fmt.Sprintf(embedDescriptionFormat, signature, markdown, examplesStr), moreSignature, moreComment
}
