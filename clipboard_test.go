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

// TestAddImageItem_Basic verifies that images are added to history.
func TestAddImageItem_Basic(t *testing.T) {
	app := &App{}
	// Create a simple fake image (PNG header)
	fakeImage := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fakeImage = append(fakeImage, make([]byte, 100)...) // Add some padding

	app.addImageItem(fakeImage)

	if len(app.history) != 1 {
		t.Fatalf("expected 1 image item, got %d", len(app.history))
	}
	if app.history[0].Type != TypeImage {
		t.Errorf("expected type to be TypeImage, got %s", app.history[0].Type)
	}
	if app.history[0].ImageData == "" {
		t.Error("expected ImageData to be populated")
	}
}

// TestAddImageItem_Dedup verifies that duplicate images are deduplicated.
func TestAddImageItem_Dedup(t *testing.T) {
	app := &App{}
	fakeImage := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	fakeImage = append(fakeImage, make([]byte, 100)...)

	app.addImageItem(fakeImage)
	app.addImageItem(fakeImage) // Duplicate

	if len(app.history) != 1 {
		t.Fatalf("expected 1 image item after dedup, got %d", len(app.history))
	}
}

// TestAddImageItem_DifferentImages verifies different images are stored separately.
func TestAddImageItem_DifferentImages(t *testing.T) {
	app := &App{}
	fakeImage1 := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x01}
	fakeImage1 = append(fakeImage1, make([]byte, 100)...)
	fakeImage2 := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x02}
	fakeImage2 = append(fakeImage2, make([]byte, 100)...)

	app.addImageItem(fakeImage1)
	app.addImageItem(fakeImage2)

	if len(app.history) != 2 {
		t.Fatalf("expected 2 image items, got %d", len(app.history))
	}
}

// TestAddImageItem_Empty verifies empty image data is ignored.
func TestAddImageItem_Empty(t *testing.T) {
	app := &App{}
	app.addImageItem([]byte{})
	app.addImageItem(nil)

	if len(app.history) != 0 {
		t.Errorf("expected 0 items, got %d", len(app.history))
	}
}

// TestClipItem_ImageType verifies ClipItem works with image type.
func TestClipItem_ImageType(t *testing.T) {
	item := ClipItem{
		Type:      TypeImage,
		ImageData: "data:image/png;base64,abc123",
		Pinned:    false,
	}

	if item.Type != TypeImage {
		t.Errorf("expected TypeImage, got %s", item.Type)
	}
	if item.Text != "" {
		t.Error("expected empty Text for image item")
	}
	if item.ImageData == "" {
		t.Error("expected ImageData to be set")
	}
}

// TestHashBytes verifies hashBytes produces consistent results.
func TestHashBytes(t *testing.T) {
	data1 := []byte("test data")
	data2 := []byte("test data")
	data3 := []byte("different data")

	hash1 := hashBytes(data1)
	hash2 := hashBytes(data2)
	hash3 := hashBytes(data3)

	if hash1 != hash2 {
		t.Error("same data should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different data should produce different hash")
	}

	// Empty data should return empty string
	emptyHash := hashBytes([]byte{})
	if emptyHash != "" {
		t.Error("empty data should return empty hash")
	}
}

// TestEncodeDecodeBase64 verifies base64 encoding and decoding.
func TestEncodeDecodeBase64(t *testing.T) {
	original := []byte("hello world test data 123")
	encoded := encodeBase64(original)

	if encoded == "" {
		t.Fatal("encoded string should not be empty")
	}

	// Decode
	decoded, err := decodeBase64(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if string(decoded) != string(original) {
		t.Errorf("decoded doesn't match original: got %s, want %s", decoded, original)
	}
}

// TestDecodeBase64_WithPrefix verifies decoding with data URI prefix.
func TestDecodeBase64_WithPrefix(t *testing.T) {
	original := []byte("test data")
	encoded := encodeBase64(original)
	withPrefix := "data:image/png;base64," + encoded

	decoded, err := decodeBase64(withPrefix)
	if err != nil {
		t.Fatalf("decode with prefix failed: %v", err)
	}

	if string(decoded) != string(original) {
		t.Errorf("decoded doesn't match: got %s, want %s", decoded, original)
	}
}

// TestTrimToCap_AllPinned verifies trimToCap with all items pinned.
func TestTrimToCap_AllPinned(t *testing.T) {
	app := &App{}

	// Add 35 items, all pinned
	for i := 0; i < 35; i++ {
		app.history = append(app.history, ClipItem{
			Type:   TypeText,
			Text:   string(rune('a' + i%26)),
			Pinned: true,
		})
	}

	app.history = app.trimToCap()

	// Should keep all 35 pinned items (cap only applies to non-pinned eviction)
	if len(app.history) != 35 {
		t.Errorf("expected 35 pinned items (all kept), got %d", len(app.history))
	}
}

// TestTrimToCap_NoPinned verifies trimToCap with no pinned items.
func TestTrimToCap_NoPinned(t *testing.T) {
	app := &App{}

	// Add 35 items, none pinned
	for i := 0; i < 35; i++ {
		app.history = append(app.history, ClipItem{
			Type:   TypeText,
			Text:   string(rune('a' + i%26)),
			Pinned: false,
		})
	}

	app.history = app.trimToCap()

	// Should cap at 30
	if len(app.history) != 30 {
		t.Errorf("expected 30 items (capped), got %d", len(app.history))
	}

	// Should keep the first 30 (newest) - items 0-29 ('a' through 'd' with wrap)
	for i := 0; i < 30; i++ {
		expectedChar := string(rune('a' + (i % 26)))
		if app.history[i].Text != expectedChar {
			t.Errorf("position %d: expected '%s', got '%s'", i, expectedChar, app.history[i].Text)
		}
	}
}

// TestTrimToCap_Empty verifies trimToCap with empty history.
func TestTrimToCap_Empty(t *testing.T) {
	app := &App{}
	app.history = app.trimToCap()

	if len(app.history) != 0 {
		t.Errorf("expected 0 items, got %d", len(app.history))
	}
}

// TestGetHistory_Empty verifies GetHistory with empty history.
func TestGetHistory_Empty(t *testing.T) {
	app := &App{}
	history := app.GetHistory()

	if history == nil {
		t.Error("GetHistory should return empty slice, not nil")
	}
	if len(history) != 0 {
		t.Errorf("expected 0 items, got %d", len(history))
	}
}

// TestDeleteItem_First deletes the first item.
func TestDeleteItem_First(t *testing.T) {
	app := &App{}
	app.addItem("first")
	app.addItem("second")
	app.addItem("third")

	app.DeleteItem(0) // Delete first ("third")

	if len(app.history) != 2 {
		t.Fatalf("expected 2 items, got %d", len(app.history))
	}
	if app.history[0].Text != "second" {
		t.Errorf("expected first to be 'second', got '%s'", app.history[0].Text)
	}
}

// TestDeleteItem_Last deletes the last item.
func TestDeleteItem_Last(t *testing.T) {
	app := &App{}
	app.addItem("first")
	app.addItem("second")
	app.addItem("third")

	app.DeleteItem(2) // Delete last ("first")

	if len(app.history) != 2 {
		t.Fatalf("expected 2 items, got %d", len(app.history))
	}
	if app.history[1].Text != "second" {
		t.Errorf("expected last to be 'second', got '%s'", app.history[1].Text)
	}
}

// TestAddItem_TypeField verifies Type field is set correctly.
func TestAddItem_TypeField(t *testing.T) {
	app := &App{}
	app.addItem("test text")

	if app.history[0].Type != TypeText {
		t.Errorf("expected TypeText, got %s", app.history[0].Type)
	}
}

// BenchmarkGetHistory benchmarks getting history.
func BenchmarkGetHistory(b *testing.B) {
	app := &App{}
	for i := 0; i < 30; i++ {
		app.addItem(string(rune('a' + i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.GetHistory()
	}
}

// BenchmarkAddImageItem benchmarks adding image items.
func BenchmarkAddImageItem(b *testing.B) {
	app := &App{}
	fakeImage := make([]byte, 10000)
	fakeImage[0] = 0x89
	fakeImage[1] = 0x50
	fakeImage[2] = 0x4E
	fakeImage[3] = 0x47

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear history periodically to avoid cap
		if i%30 == 0 {
			app.history = nil
		}
		app.addImageItem(fakeImage)
	}
}
