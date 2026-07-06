package dto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseChatPresets_AcceptsTypedAndLegacyShapes(t *testing.T) {
	t.Parallel()

	require.Equal(t, []ChatPreset{
		{ID: "a", Name: "Typed", URL: "https://chat.example.com/?k={key}"},
	}, ParseChatPresets(`[{"id":"a","name":" Typed ","url":" https://chat.example.com/?k={key} "}]`))

	require.Equal(t, []ChatPreset{
		{Name: "Legacy", URL: "fluent://chat?token={key}"},
	}, ParseChatPresets(`[{"Legacy":" fluent://chat?token={key} "}]`))
}

func TestMarshalChatPresets_StoresLegacyCompatibleShape(t *testing.T) {
	t.Parallel()

	raw, err := MarshalChatPresets([]ChatPreset{
		{Name: " Cherry ", URL: " https://cherry.example.com?apiKey={key} "},
		{Name: "", URL: "https://ignored.example.com"},
	})
	require.NoError(t, err)
	require.Equal(t, `[{"Cherry":"https://cherry.example.com?apiKey={key}"}]`, raw)
}
