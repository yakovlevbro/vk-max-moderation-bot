package callbacks

import (
	"context"
	"fmt"
	"max-moderation-bot/internal/messages"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func (h *CallbackHandler) handleBroadcastMenu(ctx context.Context, userID int64, page int) {
	if page < 1 {
		page = 1
	}

	if err := h.userStateRepo.ClearState(userID); err != nil {
		h.logger.Warn("Failed to clear state in broadcast menu", "error", err)
	}

	chats, total, err := h.svc.GetManagedChatsPaginated(ctx, userID, page)
	if err != nil {
		h.logger.Error("Failed to get managed chats for broadcast", "error", err)
		return
	}

	if len(chats) == 0 && page == 1 {
		kb := h.bot.Messages.NewKeyboardBuilder()
		kb.AddRow().AddCallback(messages.BtnBack, schemes.DEFAULT, "main_menu")
		msg := maxbot.NewMessage()
		msg.SetUser(userID)
		msg.SetText(messages.MsgNoManagedGroups)
		msg.SetFormat("markdown")
		if err := h.bot.Messages.Send(ctx, msg); err != nil {
			h.logger.Error("Failed to send no groups message in broadcast", "error", err)
		}
		return
	}

	selections, err := h.svc.GetBroadcastSelections(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get broadcast selections", "error", err)
		return
	}

	selected := make(map[int64]struct{}, len(selections))
	for _, id := range selections {
		selected[id] = struct{}{}
	}

	totalPages := (int(total) + 9) / 10

	kb := h.bot.Messages.NewKeyboardBuilder()
	for _, id := range chats {
		label := fmt.Sprintf(messages.MsgGroupDefaultLabel, id)
		if chat, err := h.bot.Chats.GetChat(ctx, id); err == nil && chat.Title != "" {
			label = chat.Title
		}
		checkmark := "☐"
		if _, ok := selected[id]; ok {
			checkmark = "✅"
		}
		kb.AddRow().AddCallback(
			fmt.Sprintf("%s %s", checkmark, label),
			schemes.DEFAULT,
			fmt.Sprintf("broadcast_toggle_%d_%d", id, page),
		)
	}

	actionRow := kb.AddRow()
	actionRow.AddCallback(messages.BtnBroadcastSelectAll, schemes.DEFAULT, fmt.Sprintf("broadcast_select_all_%d", page))
	if len(selections) > 0 {
		actionRow.AddCallback(messages.BtnBroadcastClearAll, schemes.NEGATIVE, fmt.Sprintf("broadcast_clear_all_%d", page))
	}

	if totalPages > 1 {
		navRow := kb.AddRow()
		if page > 1 {
			navRow.AddCallback(messages.BtnPrevPage, schemes.DEFAULT, fmt.Sprintf("broadcast_menu_%d", page-1))
		}
		if page < totalPages {
			navRow.AddCallback(messages.BtnNextPage, schemes.DEFAULT, fmt.Sprintf("broadcast_menu_%d", page+1))
		}
	}

	if len(selections) > 0 {
		kb.AddRow().AddCallback(
			fmt.Sprintf(messages.BtnBroadcastSend, len(selections)),
			schemes.POSITIVE,
			"broadcast_prompt",
		)
	}

	kb.AddRow().AddCallback(messages.BtnBack, schemes.DEFAULT, "main_menu")

	msg := maxbot.NewMessage()
	msg.SetUser(userID)
	msg.SetText(fmt.Sprintf(messages.MsgBroadcastSelectTitle, page, totalPages))
	msg.SetFormat("markdown")
	msg.AddKeyboard(kb)
	if err := h.bot.Messages.Send(ctx, msg); err != nil {
		h.logger.Error("Failed to send broadcast menu", "error", err)
	}
}

func (h *CallbackHandler) handleBroadcastToggle(ctx context.Context, userID, chatID int64, page int) {
	if !h.verifyAccess(ctx, userID, chatID) {
		h.logger.Warn("Access denied for broadcast toggle", "user_id", userID, "chat_id", chatID)
		return
	}
	if _, err := h.svc.ToggleBroadcastSelection(ctx, userID, chatID); err != nil {
		h.logger.Error("Failed to toggle broadcast selection", "error", err)
		return
	}
	h.handleBroadcastMenu(ctx, userID, page)
}

func (h *CallbackHandler) handleBroadcastSelectAll(ctx context.Context, userID int64, page int) {
	if err := h.svc.SelectAllChatsForBroadcast(ctx, userID); err != nil {
		h.logger.Error("Failed to select all chats for broadcast", "error", err)
		return
	}
	h.handleBroadcastMenu(ctx, userID, page)
}

func (h *CallbackHandler) handleBroadcastClearAll(ctx context.Context, userID int64, page int) {
	if err := h.svc.ClearBroadcastSelections(ctx, userID); err != nil {
		h.logger.Error("Failed to clear broadcast selections", "error", err)
		return
	}
	h.handleBroadcastMenu(ctx, userID, page)
}

func (h *CallbackHandler) handleBroadcastPrompt(ctx context.Context, userID int64) {
	selections, err := h.svc.GetBroadcastSelections(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get broadcast selections for prompt", "error", err)
		return
	}
	if len(selections) == 0 {
		h.handleBroadcastMenu(ctx, userID, 1)
		return
	}
	if err := h.userStateRepo.SetState(userID, 0, "broadcast_text"); err != nil {
		h.logger.Error("Failed to set broadcast_text state", "error", err)
		return
	}
	kb := h.bot.Messages.NewKeyboardBuilder()
	kb.AddRow().AddCallback(messages.BtnBack, schemes.DEFAULT, "broadcast_menu")
	msg := maxbot.NewMessage()
	msg.SetUser(userID)
	msg.SetText(fmt.Sprintf(messages.MsgBroadcastPromptText, len(selections)))
	msg.SetFormat("markdown")
	msg.AddKeyboard(kb)
	if err := h.bot.Messages.Send(ctx, msg); err != nil {
		h.logger.Error("Failed to send broadcast prompt", "error", err)
	}
}
