package characters

type Character struct {
	Name             string
	System           string
	Bio              []string
	Lore             []string
	Style            StyleGuide
	Topics           []string
	Goals            []Goal
	PriorityAccounts []Account
	Preferences      map[string]float64
}

type Account struct {
	Platform string
	ID       string
}

type StyleGuide struct {
	Tone        []string
	Constraints []string
}
