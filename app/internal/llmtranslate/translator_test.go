package llmtranslate

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestTranslate(t *testing.T) {
	ctx := context.Background()

	tcs := []struct {
		name        string
		input       string
		prevContext []string
		postContext []string
		language    string
		wantErr     bool
		want        string
	}{
		{
			name:     "Translate English to Spanish",
			input:    "Hello, world!",
			language: LangEs,
			wantErr:  false,
			want:     "Hola, mundo!",
		},
		{
			name: "Translate English to Spanish2",
			prevContext: []string{
				"An area also located near the Slain Theocracy.",
				`It's a symbol of strength as well as the suffering\Nthey face at the hands of their neighbors.`,
				"Orlando, it's time to change shifts.",
				"Where's your report?",
				"My apologies, Sergeant Baraja.",
			},
			input: "My thoughts were elsewhere.",
			postContext: []string{
				"But there are no changes to report today.",
				"I wish I could see as clearly in the dark as you can.",
				"I'd love to join you on the Night Watch someday, sir.",
			},
			language: LangEs,
			wantErr:  false,
			want:     "Mis pensamientos estaban en otro lugar.",
		},

		{
			name: "Translate English to Spanish3",
			prevContext: []string{
				"I've come here to make your kingdom a living hell.",
				"Wails, maledictions, and cries of pain will echo without end.",
				`It will become a carnival of suffering.`,
				"You think we would ever allow such a thing?",
				"This is not simply the Sacred Kingdom's first line of defense.",
				"It is the only one that is required!",
			},
			input: "This wall guarantees the peace of every last person who dwells behind it!",
			postContext: []string{
				"It shall not be shaken, even before you!",
				"Let's find out, shall we?",
				"Now, allow me to offer you a gift in return.",
				"Tenth-tier magic, Meteor Fall!",
			},
			language: LangEs,
			wantErr:  false,
			want:     "Esta muralla garantiza la paz de cada último hombre que vive detrás de ella.",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			translator, err := NewTranslator(ModelLlama32, "")
			if (err != nil) != tc.wantErr {
				t.Errorf("NewTranslator() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			result, err := translator.Translate(ctx, tc.prevContext, tc.postContext, tc.input, tc.language)
			if (err != nil) != tc.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if result == "" {
				t.Errorf("Translate() result is empty")
			}

			if diff := cmp.Diff(tc.want, result); diff != "" {
				t.Errorf("Mismatch (-expected +actual):\n%s", diff)
			}
		})
	}
}
