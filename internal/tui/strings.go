package tui

import "github.com/ruben/tales_vault_system/internal/profile"

type uiText struct {
	title                string
	subtitle             string
	helpTitle            string
	helpBody             []string
	helpDismiss          string
	password             string
	entropy              string
	step                 string
	block                string
	letter               string
	noWords              string
	noWordsHint          string
	currentWord          string
	pickerHint           string
	pickerHelp           string
	uniqueOn             string
	uniqueOff            string
	help                 string
	searchHelp           string
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
	searchEmpty          string
	searchCount          string
	setupTitle           string
	setupSubtitle        string
	setupHelp            string
	settingsBarPrefix    string
	settingsTitle        string
	settingsLanguage     string
	settingsPhrases      string
	settingsAllowRepeat  string
	settingsAllowAdjacent string
	settingsHelp         string
	settingsOverlayHelp  string
	settingsRegenPassword string
	settingsRegenWords   string
	settingsModalDismiss    string
	settingsFocusHint    string
	settingsListHelp     string
	settingsPickHelp     string
	settingsConfirmTitle string
	settingsConfirmApply string
	settingsConfirmDiscard string
	settingsConfirmHelp  string
	settingsPendingHint  string
	phraseOptionSummary  string
	phraseUseCase3       string
	phraseUseCase4       string
	phraseUseCase5       string
	phraseUseCase6       string
	phraseUseCase8       string
	entropyBitsUnit      string
	entropyAdjacent      string
	entropyFormula       string
	yes                  string
	no                   string
}

func uiFor(p profile.Profile) uiText {
	if p.Language == "es" {
		return uiText{
			title:                 "Tales Vault System",
			subtitle:                "Elige palabras cuya inicial coincida con cada letra de la contraseña.",
			helpTitle:               "Tales Vault System",
			helpBody: []string{
				"Convierte contraseñas en frases mnemotécnicas fáciles de recordar.",
				"Cada letra de la contraseña se asocia a una palabra cuya inicial coincide, siguiendo una plantilla gramatical (sustantivo, adjetivo, verbo…).",
				"Navega entre casillas (←→) y frases (↑↓). Pulsa Enter para abrir el rodillo y elegir palabra. Los ajustes (Tab) permiten cambiar idioma y opciones de generación.",
				"Cuando termines, pulsa g para ver el resultado final y las frases listas para memorizar.",
			},
			helpDismiss:             "F1 / Enter / Esc · cerrar",
			password:                "Contraseña",
			entropy:                 "Entropía",
			step:                    "Paso %d/%d",
			block:                   "Bloque %d/%d",
			letter:                  "letra",
			noWords:                 "No quedan palabras disponibles para esta casilla.",
			noWordsHint:             "Prueba otra casilla (←/→), otro bloque (↑/↓) o reinicia con --allow-repeat.",
			currentWord:             "Palabra",
			pickerHint:              "Enter / doble clic / rueda · cambiar palabra",
			pickerHelp:              "j/k · palabra  / · buscar  Enter · confirmar  doble clic · confirmar  Esc · cancelar",
			uniqueOn:                "No repetir palabras: sí",
			uniqueOff:               "No repetir palabras: no",
			help:                    "F1 · ayuda  ←→ · casilla  ↑↓ · frase  Enter · cambiar palabra  doble clic / rueda · cambiar  r · aleatorio  g · ver resultado  Tab/o · ajustes  Esc/q · salir",
			searchHelp:              "Escribe para filtrar · j/k · palabra  Enter · confirmar  Esc · cancelar",
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
			searchEmpty:             "Ninguna palabra coincide.",
			searchCount:             "%d coincidencias",
			setupTitle:              "Tales Vault System",
			setupSubtitle:           "Configura el idioma y las opciones de generación antes de empezar.",
			setupHelp:               "j/k · fila  Enter · editar  g · empezar  q · salir",
			settingsBarPrefix:       "Ajustes",
			settingsTitle:           "Ajustes de generación",
			settingsLanguage:        "Idioma",
			settingsPhrases:         "Frases",
			settingsAllowRepeat:     "Repetir palabras",
			settingsAllowAdjacent:   "Sin letras iguales seguidas",
			settingsHelp:            "j/k · fila  Enter · editar  Esc · salir",
			settingsListHelp:        "j/k · fila  Enter · editar  Esc · salir",
			settingsPickHelp:        "j/k · opción  Enter · confirmar  Esc · volver",
			settingsConfirmTitle:    "¿Aplicar los cambios?",
			settingsConfirmApply:    "Aplicar cambios",
			settingsConfirmDiscard:  "Descartar",
			settingsConfirmHelp:     "j/k · opción  Enter · confirmar  y/n · atajo  Esc · seguir editando",
			settingsPendingHint:     "Cambios sin aplicar · Esc para salir",
			settingsOverlayHelp:     "j/k · fila  Enter · editar  Esc · salir",
			settingsRegenPassword:   "Se regenerará la contraseña y las frases mnemotécnicas.",
			settingsRegenWords:      "Se reasignarán las palabras mnemotécnicas.",
			settingsModalDismiss:    "Esc · cerrar ajustes",
			settingsFocusHint:       "Tab / o / clic · editar ajustes",
			phraseOptionSummary:     "%d frases (%d caracteres) · ~%.0f bits",
			phraseUseCase3:          "sistemas con límite estricto de longitud o cuentas desechables",
			phraseUseCase4:          "registro en servicios en línea de uso ocasional",
			phraseUseCase5:          "uso recomendado: correo, bancos y gestores de contraseñas",
			phraseUseCase6:          "alta seguridad: clave maestra de bóveda o cuentas críticas",
			phraseUseCase8:          "máxima seguridad: secretos de muy larga vida",
			entropyBitsUnit:         "bits",
			entropyAdjacent:         "~%.1f bits (sin letras iguales seguidas)",
			entropyFormula:          "~%.1f bits (%d caracteres × log₂(%d))",
			yes:                     "sí",
			no:                      "no",
		}
	}
	return uiText{
		title:                 "Tales Vault System",
		subtitle:              "Pick words whose initials spell the password, one letter at a time.",
		helpTitle:             "Tales Vault System",
		helpBody: []string{
			"Turn passwords into memorable mnemonic phrases.",
			"Each password letter maps to a word with the same initial, following a grammatical template (noun, adjective, verb…).",
			"Move across slots (←→) and phrases (↑↓). Press Enter to open the roller and pick a word. Settings (Tab) let you change language and generation options.",
			"When you are done, press g to see the final result and phrases ready to memorize.",
		},
		helpDismiss:           "F1 / Enter / Esc · close",
		password:              "Password",
		entropy:               "Entropy",
		step:                  "Step %d/%d",
		block:                 "Block %d/%d",
		letter:                "letter",
		noWords:               "No words available for this slot.",
		noWordsHint:           "Try another slot (←/→), block (↑/↓), or restart with --allow-repeat.",
		currentWord:           "Word",
		pickerHint:            "Enter / double-click / wheel · change word",
		pickerHelp:            "j/k · word  / · search  Enter · confirm  double-click · confirm  Esc · cancel",
		uniqueOn:              "Don't repeat words: yes",
		uniqueOff:             "Don't repeat words: no",
		help:                  "F1 · help  ←→ · slot  ↑↓ · phrase  Enter · change word  double-click / wheel · change  r · random  g · show result  Tab/o · settings  Esc/q · quit",
		searchHelp:            "Type to filter · j/k · word  Enter · confirm  Esc · cancel",
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
		searchEmpty:           "No matching words.",
		searchCount:           "%d matches",
		setupTitle:            "Tales Vault System",
		setupSubtitle:         "Choose language and generation options before you start.",
		setupHelp:             "j/k · row  Enter · edit  g · start  q · quit",
		settingsBarPrefix:     "Settings",
		settingsTitle:         "Generation settings",
		settingsLanguage:      "Language",
		settingsPhrases:       "Phrases",
		settingsAllowRepeat:   "Repeat words",
		settingsAllowAdjacent: "No adjacent equal letters",
		settingsHelp:          "j/k · row  Enter · edit  Esc · exit",
		settingsListHelp:      "j/k · row  Enter · edit  Esc · exit",
		settingsPickHelp:      "j/k · option  Enter · confirm  Esc · back",
		settingsConfirmTitle:  "Apply changes?",
		settingsConfirmApply:  "Apply changes",
		settingsConfirmDiscard: "Discard",
		settingsConfirmHelp:   "j/k · option  Enter · confirm  y/n · shortcut  Esc · keep editing",
		settingsPendingHint:   "Unapplied changes · Esc to exit",
		settingsOverlayHelp:   "j/k · row  Enter · edit  Esc · exit",
		settingsRegenPassword:   "The password and mnemonic phrases will be regenerated.",
		settingsRegenWords:      "The mnemonic words will be reshuffled.",
		settingsModalDismiss:  "Esc · close settings",
		settingsFocusHint:       "Tab / o / click · edit settings",
		phraseOptionSummary:   "%d phrases (%d chars) · ~%.0f bits",
		phraseUseCase3:        "strict length limits or disposable accounts",
		phraseUseCase4:        "occasional sign-ups on online services",
		phraseUseCase5:        "recommended: email, banking, and password managers",
		phraseUseCase6:        "high security: vault master password or critical accounts",
		phraseUseCase8:        "maximum security: long-lived secrets",
		entropyBitsUnit:       "bits",
		entropyAdjacent:       "~%.1f bits (no adjacent repeats)",
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
