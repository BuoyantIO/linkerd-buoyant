package cmd

import "testing"

func TestRoot(t *testing.T) {
	t.Run("returns a valid root command", func(t *testing.T) {
		root := Root()
		err := root.Execute()
		if err != nil {
			t.Error(err)
		}
	})
}
