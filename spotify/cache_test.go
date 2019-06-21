package spotify

import (
	"context"
	"testing"
)

var ctx = context.Background()

const id = "id"

func TestAdd(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")
	checkNoError(t, n.Add(id, track.Immutable()))

	checkHasOneSong(t, n, "artist", "title")
}

func TestAddAfterReplaceAll_nil(t *testing.T) {
	n := newCache(ctx)
	checkNoError(t, n.ReplaceAll(id, nil))

	track := makeTrack("artist", "title")
	n.Add(id, track.Immutable())
	checkHasOneSong(t, n, "artist", "title")
}

func TestAddAfterReplaceAll_nilSlice(t *testing.T) {
	n := newCache(ctx)
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack(nil)))

	track := makeTrack("artist", "title")
	n.Add(id, track.Immutable())
	checkHasOneSong(t, n, "artist", "title")
}

func TestReplaceAllAfterAdd_nil(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")
	n.Add(id, track.Immutable())

	checkHasOneSong(t, n, "artist", "title")
	checkNoError(t, n.ReplaceAll(id, nil))
	checkHasZeroSong(t, n)
}

func TestReplaceAllAfterAdd_nilSlice(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")
	n.Add(id, track.Immutable())

	checkHasOneSong(t, n, "artist", "title")
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack(nil)))
	checkHasZeroSong(t, n)
}

func TestGet(t *testing.T) {
	n := newCache(ctx)
	checkHasZeroSong(t, n)
}

func TestReplaceAll(t *testing.T) {
	n := newCache(ctx)
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack{}))
	checkHasZeroSong(t, n)
}

func TestReplaceAll_nil(t *testing.T) {
	n := newCache(ctx)
	checkNoError(t, n.ReplaceAll(id, nil))
	checkHasZeroSong(t, n)
}

func TestReplaceAll_nilSlice(t *testing.T) {
	n := newCache(ctx)
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack(nil)))
	checkHasZeroSong(t, n)
}

func TestAddNil(t *testing.T) {
	n := newCache(ctx)
	err := n.Add(id, nil)
	if err == nil {
		t.Errorf("Expected error when adding nil")
	}
}

func TestReplaceWithNil(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")

	err := n.Replace(id, track.Immutable(), nil)
	if err == nil {
		t.Errorf("Expected error when replacing with nil")
	}
}

func TestReplaceNilWith(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")

	err := n.Replace(id, nil, track.Immutable())
	if err == nil {
		t.Errorf("Expected error when replacing nil")
	}
}

func TestReplace(t *testing.T) {
	n := newCache(ctx)
	track1 := makeTrack("artist1", "title1")
	track2 := makeTrack("artist2", "title2")
	track1.Id = "abc"
	track2.Id = "def"

	checkNoError(t, n.Add(id, track1.Immutable()))
	checkHasOneSong(t, n, "artist1", "title1")
	checkNoError(t, n.Replace(id, track1.Immutable(), track2.Immutable()))
	checkHasOneSong(t, n, "artist2", "title2")
}

func TestPointersAreSet(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")

	checkNoError(t, n.Add(id, track.Immutable()))
	checkHasOneSong(t, n, "artist", "title")

	// Modyfing underlying data results in cache changes
	track.Name = "abc"
	checkHasOneSong(t, n, "artist", "abc")
}

func TestPointersAreReturned(t *testing.T) {
	n := newCache(ctx)
	track := makeTrack("artist", "title")

	checkNoError(t, n.Add(id, track.Immutable()))
	checkHasOneSong(t, n, "artist", "title")

	track2 := makeTrack("another", "different")
	n.Get(id)[0] = track2.Immutable()

	// song in cache should not be modified
	checkHasOneSong(t, n, "artist", "title")
}

func checkNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf("Unexpected error: %v.", err)
	}
}

func checkHasZeroSong(t *testing.T, c Cache) {
	tracks := c.Get(id)
	if tracks == nil || len(tracks) != 0 {
		t.Errorf("cache got: %v, want: empty", tracks)
	}
}

func checkHasOneSong(t *testing.T, c Cache, artist string, title string) {
	tracks := c.Get(id)
	if len(tracks) != 1 || tracks[0].Title() != title || tracks[0].Artist() != artist {
		t.Errorf("cache got: %v, want one song: %v - %v", tracks, artist, title)
	}
}
