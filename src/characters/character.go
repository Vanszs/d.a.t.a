package characters

type Character struct {
	Name        string
	System      string
	Bio         []string
	Lore        []string
	Style       StyleGuide
	Topics      []string
	Goals       []Goal
	Preferences map[string]float64
}

type StyleGuide struct {
	Tone        []string
	Constraints []string
}
