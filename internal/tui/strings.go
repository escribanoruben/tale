package tui

import "github.com/escribanoruben/tale/internal/profile"

type uiText struct {
	commandTagline       string
	subtitle             string
	helpBody             []string
	helpShortcutsTitle   string
	helpShortcuts        []string
	helpDismiss          string
	password             string
	entropy              string
	noWords              string
	noWordsHint          string
	randomTitle          string
	randomWord           string
	randomPhrase         string
	randomAll            string
	randomCancel         string
	changeLetter         string
	changeCancel         string
	summaryTitle         string
	recovery             string
	cancelled            string
	setupSubtitle        string
	setupHelp            string
	settingsBarPrefix    string
	settingsTitle        string
	settingsLanguage     string
	settingsPhrases      string
	settingsAllowAdjacent string
	settingsAdjacentPrevent string
	settingsAdjacentAllow   string
	settingsSubtitle     string
	settingsRegenPassword string
	settingsFocusHint    string
	settingsListHelp         string
	settingsListHelpPending  string
	settingsConfirmApply     string
	settingsConfirmDiscard   string
	settingsDiscardPrompt    string
	settingsDiscardHelp      string
	settingsPendingHint  string
	phraseOptionSummary  string
	phraseUseCase3       string
	phraseUseCase4       string
	phraseUseCase5       string
	phraseUseCase6       string
	phraseUseCase8       string
	entropyBitsUnit      string
	entropyFormula       string
	yes                  string
	no                   string
}

func uiFor(p profile.Profile) uiText {
	if p.Language == "es" {
		return uiText{
			commandTagline:          "Una implementación del Tales Vault System",
			subtitle:                "Elige palabras cuya inicial coincida con cada letra de la contraseña.",
			helpBody: []string{
				"Convierte contraseñas en frases mnemotécnicas fáciles de recordar.",
				"Cada letra de la contraseña se asocia a una palabra cuya inicial coincide, siguiendo una plantilla gramatical (sustantivo, adjetivo, verbo…).",
				"Navega entre casillas (←→) y frases (↑↓). El rodillo a la izquierda muestra las palabras disponibles; usa n/p o la rueda del ratón para cambiar.",
				"Cuando termines, pulsa Enter para ver el resultado final y las frases listas para memorizar.",
			},
			helpShortcutsTitle:    "Atajos",
			helpShortcuts: []string{
				"h/l · casilla anterior / siguiente",
				"j/k · frase anterior / siguiente",
				"n/p · palabra anterior / siguiente",
				"Enter · ver resultado",
				"r · menú aleatorio",
				"Tab · ajustes",
				"? · ayuda",
				"Esc · salir",
			},
			helpDismiss:             "? / Enter / Esc · cerrar",
			password:                "Contraseña",
			entropy:                 "Entropía",
			noWords:                 "No quedan palabras disponibles para esta casilla.",
			noWordsHint:             "Prueba otra casilla (←/→) u otro bloque (↑/↓).",
			randomTitle:             "Aleatorio",
			randomWord:              "w  palabra actual",
			randomPhrase:            "f  frase del bloque",
			randomAll:               "a  toda la contraseña",
			randomCancel:            "Esc cancelar",
			changeLetter:            "Escribe una letra del alfabeto para cambiar la contraseña en esta posición:",
			changeCancel:            "Esc cancelar",
			summaryTitle:            "Frases mnemotécnicas",
			recovery:                "Recuperación: toma la primera letra de cada palabra (sin tildes, ñ→n).",
			cancelled:               "Cancelado.",
			setupSubtitle:           "Configura el idioma y las opciones de generación antes de empezar.",
			setupHelp:               "j/k · mover  Espacio · elegir  g · empezar  Esc · salir",
			settingsBarPrefix:       "Ajustes",
			settingsTitle:           "Ajustes de generación",
			settingsSubtitle:        "Elige idioma, longitud y restricciones de la contraseña.",
			settingsLanguage:        "Idioma",
			settingsPhrases:         "Frases",
			settingsAllowAdjacent:   "Sin letras iguales seguidas",
			settingsAdjacentPrevent: "Sí — evita repeticiones consecutivas; reduce typos al escribir rápido (con «rr» u otras parejas suele quedar solo una letra)",
			settingsAdjacentAllow:   "No — permite letras iguales seguidas; un poco más de entropía, pero más fácil equivocarse al teclear",
			settingsListHelp:        "j/k · mover  Espacio · elegir  1-3 · sección  Esc · volver",
			settingsListHelpPending: "j/k · mover  Espacio · elegir o aplicar  Esc · descartar",
			settingsConfirmApply:    "Aplicar cambios",
			settingsConfirmDiscard:  "Descartar y volver",
			settingsDiscardPrompt:   "¿Descartar los cambios sin aplicar?",
			settingsDiscardHelp:     "s · descartar   n / Esc · seguir editando",
			settingsPendingHint:     "Cambios sin aplicar · Esc para salir",
			settingsRegenPassword:   "Se regenerará la contraseña y las frases mnemotécnicas.",
			settingsFocusHint:       "Tab · editar ajustes  ? · ayuda",
			phraseOptionSummary:     "%d frases (%d caracteres) · ~%.0f bits",
			phraseUseCase3:          "sistemas con límite estricto de longitud o cuentas desechables",
			phraseUseCase4:          "registro en servicios en línea de uso ocasional",
			phraseUseCase5:          "uso recomendado: correo, bancos y gestores de contraseñas",
			phraseUseCase6:          "alta seguridad: clave maestra de bóveda o cuentas críticas",
			phraseUseCase8:          "máxima seguridad: secretos de muy larga vida",
			entropyBitsUnit:         "bits",
			entropyFormula:          "~%.1f bits (%d caracteres × log₂(%d))",
			yes:                     "sí",
			no:                      "no",
		}
	}
	return uiText{
		commandTagline:        "An implementation of the Tales Vault System",
		subtitle:              "Pick words whose initials spell the password, one letter at a time.",
		helpBody: []string{
			"Turn passwords into memorable mnemonic phrases.",
			"Each password letter maps to a word with the same initial, following a grammatical template (noun, adjective, verb…).",
			"Move across slots (←→) and phrases (↑↓). The roller on the left shows available words; use n/p or the mouse wheel to change.",
			"When you are done, press Enter to see the final result and phrases ready to memorize.",
		},
		helpShortcutsTitle:  "Shortcuts",
		helpShortcuts: []string{
			"h/l · previous / next slot",
			"j/k · previous / next phrase",
			"n/p · previous / next word",
			"Enter · show result",
			"r · randomize menu",
			"Tab · settings",
			"? · help",
			"Esc · quit",
		},
		helpDismiss:           "? / Enter / Esc · close",
		password:              "Password",
		entropy:               "Entropy",
		noWords:               "No words available for this slot.",
		noWordsHint:           "Try another slot (←/→) or block (↑/↓).",
		randomTitle:           "Randomize",
		randomWord:            "w  current word",
		randomPhrase:          "f  current block phrase",
		randomAll:             "a  entire mnemonic",
		randomCancel:          "Esc cancel",
		changeLetter:          "Type a letter from the alphabet to change the password at this position:",
		changeCancel:          "Esc cancel",
		summaryTitle:          "Mnemonic phrases",
		recovery:              "Recovery: take the first letter of each word (strip accents, ñ→n).",
		cancelled:             "Cancelled.",
		setupSubtitle:         "Choose language and generation options before you start.",
		setupHelp:             "j/k · move  Space · select  g · start  Esc · quit",
		settingsBarPrefix:     "Settings",
		settingsTitle:         "Generation settings",
		settingsSubtitle:      "Choose language, length, and password constraints.",
		settingsLanguage:      "Language",
		settingsPhrases:       "Phrases",
		settingsAllowAdjacent: "No adjacent equal letters",
		settingsAdjacentPrevent: "Yes — no consecutive repeats; fewer typos when typing fast (with «rr» and similar pairs you may type only one letter)",
		settingsAdjacentAllow:   "No — allow equal adjacent letters; slightly more entropy, but easier to mistype doubles when typing fast",
		settingsListHelp:      "j/k · move  Space · select  1-3 · section  Esc · back",
		settingsListHelpPending: "j/k · move  Space · select or apply  Esc · discard",
		settingsConfirmApply:  "Apply changes",
		settingsConfirmDiscard: "Discard and go back",
		settingsDiscardPrompt: "Discard unapplied changes?",
		settingsDiscardHelp:   "y · discard   n / Esc · keep editing",
		settingsPendingHint:   "Unapplied changes · Esc to exit",
		settingsRegenPassword:   "The password and mnemonic phrases will be regenerated.",
		settingsFocusHint:       "Tab · edit settings  ? · help",
		phraseOptionSummary:   "%d phrases (%d chars) · ~%.0f bits",
		phraseUseCase3:        "strict length limits or disposable accounts",
		phraseUseCase4:        "occasional sign-ups on online services",
		phraseUseCase5:        "recommended: email, banking, and password managers",
		phraseUseCase6:        "high security: vault master password or critical accounts",
		phraseUseCase8:        "maximum security: long-lived secrets",
		entropyBitsUnit:       "bits",
		entropyFormula:        "~%.1f bits (%d chars × log₂(%d))",
		yes:                   "yes",
		no:                    "no",
	}
}

func posLabel(lang, pos string) string {
	if lang == "es" {
		switch pos {
		case "noun":
			return "sustantivo"
		case "adjective":
			return "adjetivo"
		case "verb":
			return "verbo"
		}
	}
	return pos
}
