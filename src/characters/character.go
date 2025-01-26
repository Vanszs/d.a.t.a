package characters

type Character struct {
	ID               string
	Name             string
	System           string
	Bio              []string
	Lore             []string
	Style            StyleGuide
	Topics           []string
	Goals            []Goal
	PreferenceVector map[string]float64
}

type StyleGuide struct {
	Tone        []string
	Actions     []string
	Constraints []string
}

func (c *Character) Learn(outcome Outcome) {
	// Update preferences based on action outcomes
	for topic, impact := range outcome.TopicImpacts {
		c.PreferenceVector[topic] += impact * outcome.Success
	}
}
