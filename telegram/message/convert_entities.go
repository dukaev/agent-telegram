package message

import (
	"github.com/gotd/td/tg"
)

// convertReactions converts reactions to map slice.
func convertReactions(reactions tg.MessageReactions) []map[string]any {
	result := make([]map[string]any, 0, len(reactions.Results))
	for _, r := range reactions.Results {
		reaction := map[string]any{
			"count": r.Count,
		}
		if r.Reaction != nil {
			switch react := r.Reaction.(type) {
			case *tg.ReactionEmoji:
				reaction["emoticon"] = react.Emoticon
			case *tg.ReactionCustomEmoji:
				reaction["document_id"] = react.DocumentID
			}
		}
		if r.ChosenOrder != 0 {
			reaction["chosen_order"] = r.ChosenOrder
		}
		result = append(result, reaction)
	}
	return result
}

// convertEntities converts message entities to map slice.
func convertEntities(entities []tg.MessageEntityClass) []map[string]any {
	result := make([]map[string]any, 0, len(entities))
	for _, e := range entities {
		entity := map[string]any{
			"offset": e.GetOffset(),
			"length": e.GetLength(),
		}
		switch ent := e.(type) {
		case *tg.MessageEntityTextURL:
			entity["type"] = "text_url"
			entity["url"] = ent.URL
		case *tg.MessageEntityURL:
			entity["type"] = "url"
		case *tg.MessageEntityEmail:
			entity["type"] = "email"
		case *tg.MessageEntityHashtag:
			entity["type"] = "hashtag"
		case *tg.MessageEntityCashtag:
			entity["type"] = "cashtag"
		case *tg.MessageEntityMention:
			entity["type"] = "mention"
		case *tg.MessageEntityMentionName:
			entity["type"] = "mention_name"
			entity["user_id"] = ent.UserID
		case *tg.MessageEntityBotCommand:
			entity["type"] = "bot_command"
		case *tg.MessageEntityBold:
			entity["type"] = "bold"
		case *tg.MessageEntityItalic:
			entity["type"] = "italic"
		case *tg.MessageEntityUnderline:
			entity["type"] = "underline"
		case *tg.MessageEntityStrike:
			entity["type"] = "strike"
		case *tg.MessageEntityCode:
			entity["type"] = "code"
		case *tg.MessageEntityPre:
			entity["type"] = "pre"
			if ent.Language != "" {
				entity["language"] = ent.Language
			}
		case *tg.MessageEntityBlockquote:
			entity["type"] = "blockquote"
		}
		result = append(result, entity)
	}
	return result
}
