package main

import "testing"

func TestTranslatablePayloadIncludesRebuildOpportunityContent(t *testing.T) {
	pivotIdea := map[string]any{"name": "LoopWardrobe", "concept": "Circular fashion marketplace"}
	lessons := []any{"Anchor credits to real value"}
	payload := translatablePayload("rebuild_plan", map[string]any{
		"name":        "99dresses",
		"pivot_idea":  pivotIdea,
		"the_loot":    lessons,
		"source_hash": "not-translatable",
	})

	if payload["pivot_idea"] == nil || payload["the_loot"] == nil {
		t.Fatalf("rebuild opportunity content missing from translation payload: %#v", payload)
	}
	if _, exists := payload["source_hash"]; exists {
		t.Fatalf("source metadata leaked into translation payload: %#v", payload)
	}
}
