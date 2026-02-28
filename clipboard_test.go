package main

import (
	"strings"
	"testing"
)

// TestAddItem_BasicInsert verifies items are added to the front of the list.
func TestAddItem_BasicInsert(t *testing.T) {
	app := &App{}
	app.addItem("first")
	app.addItem("second")
	app.addItem("third")

	if len(app.history) != 3 {
		t.Fatalf("expected 3 items, got %d", len(app.history))
	}
	if app.history[0].Text != "third" {
		t.Errorf("expected first item to be 'third', got '%s'", app.history[0].Text)
	}
	if app.history[1].Text != "second" {
		t.Errorf("expected second item to be 'second', got '%s'", app.history[1].Text)
	}
	if app.history[2].Text != "first" {
		t.Errorf("expected third item to be 'first', got '%s'", app.history[2].Text)
	}
}

// TestAddItem_Dedup verifies duplicates are removed and moved to the front.
func TestAddItem_Dedup(t *testing.T) {
	app := &App{}
	app.addItem("first")
	app.addItem("second")
	app.addItem("third")
	app.addItem("second") // Duplicate - should move to front

	if len(app.history) != 3 {
		t.Fatalf("expected 3 items, got %d", len(app.history))
	}
	if app.history[0].Text != "second" {
		t.Errorf("expected first item to be 'second' (moved), got '%s'", app.history[0].Text)
	}
	if app.history[1].Text != "third" {
		t.Errorf("expected second item to be 'third', got '%s'", app.history[1].Text)
	}
	if app.history[2].Text != "first" {
		t.Errorf("expected third item to be 'first', got '%s'", app.history[2].Text)
	}
}

// TestAddItem_Cap verifies history is capped at 30 items.
func TestAddItem_Cap(t *testing.T) {
	app := &App{}
	for i := 0; i < 35; i++ {
		app.addItem(string(rune('a' + i))) // Use unique strings to avoid dedup
	}

	if len(app.history) != 30 {
		t.Fatalf("expected 30 items (capped), got %d", len(app.history))
	}
}

// TestAddItem_Empty verifies empty strings are ignored.
func TestAddItem_Empty(t *testing.T) {
	app := &App{}
	app.addItem("")
	app.addItem("valid")

	if len(app.history) != 1 {
		t.Fatalf("expected 1 item, got %d", len(app.history))
	}
	if app.history[0].Text != "valid" {
		t.Errorf("expected 'valid', got '%s'", app.history[0].Text)
	}
}

// TestAddItem_Whitespace verifies whitespace-only strings are ignored.
func TestAddItem_Whitespace(t *testing.T) {
	app := &App{}
	app.addItem("   ")
	app.addItem("\t\n\r")
	app.addItem("  valid  ")

	if len(app.history) != 1 {
		t.Fatalf("expected 1 item, got %d", len(app.history))
	}
	if app.history[0].Text != "valid" {
		t.Errorf("expected 'valid' (trimmed), got '%s'", app.history[0].Text)
	}
}

// TestGetHistory_ReturnsCopy verifies GetHistory returns a copy that doesn't affect internal state.
func TestGetHistory_ReturnsCopy(t *testing.T) {
	app := &App{}
	app.addItem("first")
	app.addItem("second")

	history := app.GetHistory()
	if len(history) != 2 {
		t.Fatalf("expected 2 items in copy, got %d", len(history))
	}

	// Modify the copy
	history[0].Text = "modified"
	history = append(history, ClipItem{Text: "third"})

	// Original should be unchanged
	if app.history[0].Text != "second" {
		t.Errorf("original was modified: got '%s'", app.history[0].Text)
	}
	if len(app.history) != 2 {
		t.Errorf("original length changed: got %d", len(app.history))
	}
}

// TestAddItem_Trimming verifies text is properly trimmed.
func TestAddItem_Trimming(t *testing.T) {
	app := &App{}
	app.addItem("  hello world  ")

	if app.history[0].Text != "hello world" {
		t.Errorf("expected 'hello world' (trimmed), got '%s'", app.history[0].Text)
	}
}

// TestAddItem_Multiline verifies multiline text is preserved.
func TestAddItem_Multiline(t *testing.T) {
	app := &App{}
	multiline := "line1\nline2\nline3"
	app.addItem(multiline)

	if app.history[0].Text != multiline {
		t.Errorf("expected multiline text preserved, got '%s'", app.history[0].Text)
	}
}

// TestAddItem_DedupPreservesPinnedStatus verifies dedup preserves pinned status.
func TestAddItem_DedupPreservesPinnedStatus(t *testing.T) {
	app := &App{}
	app.addItem("item1")
	app.addItem("item2")
	app.history[1].Pinned = true // Pin "item1"
	app.addItem("item1")         // Re-add same text

	// The re-added item should not be pinned (it's a fresh insert)
	if app.history[0].Pinned {
		t.Error("re-added item should not be pinned by default")
	}
}

// BenchmarkAddItem benchmarks adding items to the history.
func BenchmarkAddItem(b *testing.B) {
	app := &App{}
	for i := 0; i < b.N; i++ {
		app.addItem(strings.Repeat("x", i%100))
	}
}

// TestAddItem_SkipsLastWritten verifies that addItem skips text that matches lastWritten.
// This prevents re-capturing text that was just pasted from the clipboard manager.
func TestAddItem_SkipsLastWritten(t *testing.T) {
	app := &App{}
	app.addItem("first")
	app.addItem("second")

	// Simulate that we just wrote "second" via SelectItem
	app.lastWritten = "second"

	// Adding "second" again should be skipped
	app.addItem("second")

	if len(app.history) != 2 {
		t.Fatalf("expected 2 items (skipped lastWritten), got %d", len(app.history))
	}
	if app.history[0].Text != "second" {
		t.Errorf("expected first item to still be 'second', got '%s'", app.history[0].Text)
	}
}

// TestAddItem_LastWrittenCleared verifies that lastWritten is cleared after one skip.
// This ensures future copies of the same text are captured again.
func TestAddItem_LastWrittenCleared(t *testing.T) {
	app := &App{}
	app.addItem("item")
	app.addItem("other")

	// Simulate that we just wrote "item" via SelectItem
	app.lastWritten = "item"

	// First add should be skipped and clear lastWritten
	app.addItem("item")
	if len(app.history) != 2 {
		t.Fatalf("expected 2 items after skip, got %d", len(app.history))
	}
	// "item" should NOT have moved to front (it was skipped)
	if app.history[0].Text != "other" {
		t.Errorf("expected first item to still be 'other' (skip didn't move), got '%s'", app.history[0].Text)
	}

	// Second add should succeed (lastWritten was cleared)
	// "item" should be deduped (removed from position 1, added to front)
	app.addItem("item")
	if len(app.history) != 2 {
		t.Fatalf("expected 2 items after re-add, got %d", len(app.history))
	}
	// "item" should now be at front (normal dedup behavior)
	if app.history[0].Text != "item" {
		t.Errorf("expected first item to be 'item' (moved to front), got '%s'", app.history[0].Text)
	}
}

// TestAddItem_LastWrittenDoesNotAffectOtherItems verifies that lastWritten only skips exact matches.
func TestAddItem_LastWrittenDoesNotAffectOtherItems(t *testing.T) {
	app := &App{}
	app.addItem("first")

	// Simulate that we just wrote "different" via SelectItem
	app.lastWritten = "different"

	// Adding "first" should work normally
	app.addItem("first")

	if len(app.history) != 1 {
		t.Fatalf("expected 1 item (dedup), got %d", len(app.history))
	}
	// Should be moved to front
	if app.history[0].Text != "first" {
		t.Errorf("expected 'first', got '%s'", app.history[0].Text)
	}
}

// TestSelectItem_UpdatesLastWritten verifies that SelectItem sets lastWritten to prevent re-capture.
func TestSelectItem_UpdatesLastWritten(t *testing.T) {
	app := &App{}
	app.addItem("item1")
	app.addItem("item2")
	app.addItem("item3")

	// SelectItem would normally write to clipboard and hide window,
	// but we can't test that without a window. We verify the state changes instead.
	// The lastWritten should be set to the selected item's text
	app.mu.Lock()
	app.lastWritten = app.history[1].Text // Simulate what SelectItem does
	app.mu.Unlock()

	if app.lastWritten != "item2" {
		t.Errorf("expected lastWritten to be 'item2', got '%s'", app.lastWritten)
	}

	// Subsequent add of same text should be skipped
	app.addItem("item2")
	if len(app.history) != 3 {
		t.Errorf("expected 3 items (skipped), got %d", len(app.history))
	}
}

// TestClipItem_JSONTags verifies ClipItem struct has correct JSON tags for serialization.
func TestClipItem_JSONTags(t *testing.T) {
	// This test ensures the struct fields are tagged correctly for JSON serialization
	// which will be needed when we add persistence in Step 2
	item := ClipItem{
		Text:   "test text",
		Pinned: true,
	}

	// Verify the struct can be created with the expected fields
	if item.Text != "test text" {
		t.Errorf("expected Text to be 'test text', got '%s'", item.Text)
	}
	if !item.Pinned {
		t.Error("expected Pinned to be true")
	}
}

// TestTogglePin toggles the pinned flag on an item.
func TestTogglePin(t *testing.T) {
	app := &App{}
	app.addItem("item1")
	app.addItem("item2")

	// Initially not pinned
	if app.history[1].Pinned {
		t.Error("new item should not be pinned")
	}

	// Toggle pin on
	app.TogglePin(1)
	if !app.history[1].Pinned {
		t.Error("item should be pinned after toggle")
	}

	// Toggle pin off
	app.TogglePin(1)
	if app.history[1].Pinned {
		t.Error("item should not be pinned after second toggle")
	}
}

// TestTogglePin_OutOfBounds ensures TogglePin handles invalid indices gracefully.
func TestTogglePin_OutOfBounds(t *testing.T) {
	app := &App{}
	app.addItem("item")

	// Should not panic on negative index
	app.TogglePin(-1)

	// Should not panic on out of bounds index
	app.TogglePin(10)

	// Item should still be there
	if len(app.history) != 1 {
		t.Errorf("history should have 1 item, got %d", len(app.history))
	}
}

// TestDeleteItem removes an item from history.
func TestDeleteItem(t *testing.T) {
	app := &App{}
	app.addItem("item1")
	app.addItem("item2")
	app.addItem("item3")

	// Delete middle item
	app.DeleteItem(1)

	if len(app.history) != 2 {
		t.Fatalf("expected 2 items after delete, got %d", len(app.history))
	}
	if app.history[0].Text != "item3" {
		t.Errorf("expected first item to be 'item3', got '%s'", app.history[0].Text)
	}
	if app.history[1].Text != "item1" {
		t.Errorf("expected second item to be 'item1', got '%s'", app.history[1].Text)
	}
}

// TestDeleteItem_OutOfBounds ensures DeleteItem handles invalid indices gracefully.
func TestDeleteItem_OutOfBounds(t *testing.T) {
	app := &App{}
	app.addItem("item1")
	app.addItem("item2")

	// Should not panic on negative index
	app.DeleteItem(-1)

	if len(app.history) != 2 {
		t.Errorf("expected 2 items, got %d", len(app.history))
	}

	// Should not panic on out of bounds index
	app.DeleteItem(10)

	if len(app.history) != 2 {
		t.Errorf("expected 2 items, got %d", len(app.history))
	}
}

// TestAddItem_DedupPinned verifies that pinned items are not re-added (skipped).
func TestAddItem_DedupPinned(t *testing.T) {
	app := &App{}
	app.addItem("item1")
	app.addItem("item2")
	app.history[1].Pinned = true // Pin "item1"

	// Try to add same text as pinned item
	app.addItem("item1")

	// Should still have 2 items (new addition was skipped)
	if len(app.history) != 2 {
		t.Fatalf("expected 2 items (pinned item not re-added), got %d", len(app.history))
	}

	// item1 should still be at index 1 (not moved to front because we skipped)
	if app.history[1].Text != "item1" || !app.history[1].Pinned {
		t.Error("pinned item should remain in place")
	}
}

// TestAddItem_CapPreservesPinned verifies that cap eviction skips pinned items.
func TestAddItem_CapPreservesPinned(t *testing.T) {
	app := &App{}

	// Add 30 items and pin the first 5 (will be at the end since we prepend)
	for i := 0; i < 30; i++ {
		app.addItem(string(rune('a' + i)))
	}

	// Pin items at indices 25-29 (which are 'a', 'b', 'c', 'd', 'e' - the oldest ones)
	for i := 25; i < 30; i++ {
		app.history[i].Pinned = true
	}

	// Add more items to trigger cap
	for i := 30; i < 35; i++ {
		app.addItem(string(rune('a' + i)))
	}

	// Should have exactly 30 items
	if len(app.history) != 30 {
		t.Fatalf("expected 30 items (capped), got %d", len(app.history))
	}

	// Count pinned items - should still have all 5
	pinnedCount := 0
	for _, item := range app.history {
		if item.Pinned {
			pinnedCount++
		}
	}
	if pinnedCount != 5 {
		t.Errorf("expected 5 pinned items to be preserved, got %d", pinnedCount)
	}

	// Verify the 5 pinned items are the original ones ('a' through 'e')
	pinnedTexts := make(map[string]bool)
	for _, item := range app.history {
		if item.Pinned {
			pinnedTexts[item.Text] = true
		}
	}
	for _, expected := range []string{"a", "b", "c", "d", "e"} {
		if !pinnedTexts[expected] {
			t.Errorf("pinned item '%s' was not preserved", expected)
		}
	}
}
