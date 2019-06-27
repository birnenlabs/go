package spotify

import (
	"testing"
)

const id = "id"

func TestAdd(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")
	checkNoError(t, n.Add(id, track.immutable()))

	checkHasOneSong(t, n, "artist", "title")
}

func TestAddAfterReplaceAll_nil(t *testing.T) {
	n := newCache()
	checkNoError(t, n.ReplaceAll(id, nil))

	track := makeTrack("artist", "title")
	n.Add(id, track.immutable())
	checkHasOneSong(t, n, "artist", "title")
}

func TestAddAfterReplaceAll_nilSlice(t *testing.T) {
	n := newCache()
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack(nil)))

	track := makeTrack("artist", "title")
	n.Add(id, track.immutable())
	checkHasOneSong(t, n, "artist", "title")
}

func TestReplaceAllAfterAdd_nil(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")
	n.Add(id, track.immutable())

	checkHasOneSong(t, n, "artist", "title")
	checkNoError(t, n.ReplaceAll(id, nil))
	checkHasZeroSong(t, n)
}

func TestReplaceAllAfterAdd_nilSlice(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")
	n.Add(id, track.immutable())

	checkHasOneSong(t, n, "artist", "title")
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack(nil)))
	checkHasZeroSong(t, n)
}

func TestGet(t *testing.T) {
	n := newCache()
	checkHasZeroSong(t, n)
}

func TestReplaceAll(t *testing.T) {
	n := newCache()
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack{}))
	checkHasZeroSong(t, n)
}

func TestReplaceAll_nil(t *testing.T) {
	n := newCache()
	checkNoError(t, n.ReplaceAll(id, nil))
	checkHasZeroSong(t, n)
}

func TestReplaceAll_nilSlice(t *testing.T) {
	n := newCache()
	checkNoError(t, n.ReplaceAll(id, []*ImmutableSpotifyTrack(nil)))
	checkHasZeroSong(t, n)
}

func TestAddNil(t *testing.T) {
	n := newCache()
	err := n.Add(id, nil)
	if err == nil {
		t.Errorf("Expected error when adding nil")
	}
}

func TestRemove(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")

	checkNoError(t, n.Add(id, track.immutable()))
	checkHasOneSong(t, n, "artist", "title")

	checkNoError(t, n.Remove(id, track.immutable()))
	checkHasZeroSong(t, n)
}

func TestRemoveNil(t *testing.T) {
	n := newCache()

	err := n.Remove(id, nil)
	if err == nil {
		t.Errorf("Expected error when replacing nil")
	}
}

func TestRemoveNonExisting(t *testing.T) {
	n := newCache()
	track1 := makeTrack("artist", "title")
	track2 := makeTrack("artist2", "title2")

	checkNoError(t, n.Add(id, track1.immutable()))

	err := n.Remove(id, track2.immutable())
	if err == nil {
		t.Errorf("Expected error when replacing non existing")
	}
}

func TestRemove_beginning(t *testing.T) {
	n := newCache()
	track1 := makeTrack("a1", "t1")
	track2 := makeTrack("a2", "t2")
	track3 := makeTrack("a3", "t3")

	checkNoError(t, n.Add(id, track1.immutable()))
	checkNoError(t, n.Add(id, track2.immutable()))
	checkNoError(t, n.Add(id, track3.immutable()))
	checkNoError(t, n.Remove(id, track1.immutable()))

	checkHasTwoSongs(t, n, "a2", "t2", "a3", "t3")
}

func TestRemove_middle(t *testing.T) {
	n := newCache()
	track1 := makeTrack("a1", "t1")
	track2 := makeTrack("a2", "t2")
	track3 := makeTrack("a3", "t3")

	checkNoError(t, n.Add(id, track1.immutable()))
	checkNoError(t, n.Add(id, track2.immutable()))
	checkNoError(t, n.Add(id, track3.immutable()))
	checkNoError(t, n.Remove(id, track2.immutable()))

	checkHasTwoSongs(t, n, "a1", "t1", "a3", "t3")
}

func TestRemove_end(t *testing.T) {
	n := newCache()
	track1 := makeTrack("a1", "t1")
	track2 := makeTrack("a2", "t2")
	track3 := makeTrack("a3", "t3")

	checkNoError(t, n.Add(id, track1.immutable()))
	checkNoError(t, n.Add(id, track2.immutable()))
	checkNoError(t, n.Add(id, track3.immutable()))
	checkNoError(t, n.Remove(id, track3.immutable()))

	checkHasTwoSongs(t, n, "a1", "t1", "a2", "t2")
}

func TestRemoveAllDuplicates(t *testing.T) {
	n := newCache()
	track := makeTrack("a", "t")

	checkNoError(t, n.Add(id, track.immutable()))
	checkNoError(t, n.Add(id, track.immutable()))
	checkHasTwoSongs(t, n, "a", "t", "a", "t")

	n.Remove(id, track.immutable())
	checkHasZeroSong(t, n)
}

func TestImmutableIsCopied(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")

	checkNoError(t, n.Add(id, track.immutable()))
	checkHasOneSong(t, n, "artist", "title")

	// Modyfing underlying data results in cache not changed
	track.Name = "abc"
	checkHasOneSong(t, n, "artist", "title")
}

func TestCannotModifyCacheArray(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")

	checkNoError(t, n.Add(id, track.immutable()))
	checkHasOneSong(t, n, "artist", "title")

	track2 := makeTrack("another", "different")
	n.Get(id)[0] = track2.immutable()

	// song in cache should not be modified
	checkHasOneSong(t, n, "artist", "title")
}

func TestCannotModifyCacheArrayAfterReplaceAll(t *testing.T) {
	n := newCache()
	track := makeTrack("artist", "title")

	all := append([]*ImmutableSpotifyTrack{}, track.immutable())
	checkNoError(t, n.ReplaceAll(id, all))
	checkHasOneSong(t, n, "artist", "title")

	newTrack := makeTrack("other", "other")
	all[0] = newTrack.immutable()
	// song in cache should not be modified
	checkHasOneSong(t, n, "artist", "title")
}

func TestMoreSongs(t *testing.T) {
	n := newCache()
	track1 := makeTrack("a1", "t1")
	track2 := makeTrack("a2", "t2")
	track3 := makeTrack("a3", "t3")

	checkNoError(t, n.Add(id, track1.immutable()))
	checkNoError(t, n.Add(id, track2.immutable()))

	checkHasTwoSongs(t, n, "a1", "t1", "a2", "t2")

	checkNoError(t, n.Remove(id, track2.immutable()))
	checkNoError(t, n.Add(id, track3.immutable()))
	checkHasTwoSongs(t, n, "a1", "t1", "a3", "t3")
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
	if c.IsCached(id) {
		t.Errorf("expectind IsCached == false for 0 songs")
	}
}

func checkHasOneSong(t *testing.T, c Cache, artist string, title string) {
	tracks := c.Get(id)
	if len(tracks) != 1 || tracks[0].Title() != title || tracks[0].Artist() != artist {
		t.Errorf("cache got: %v, want one song: %v - %v", tracks, artist, title)
	}
	if !c.IsCached(id) {
		t.Errorf("expectind IsCached == true for 1 song")
	}
}

func checkHasTwoSongs(t *testing.T, c Cache, artist1 string, title1 string, artist2 string, title2 string) {
	tracks := c.Get(id)
	if len(tracks) != 2 || tracks[0].Title() != title1 || tracks[0].Artist() != artist1 || tracks[1].Title() != title2 || tracks[1].Artist() != artist2 {
		t.Errorf("cache got: %v, want: %v - %v, %v - %v", tracks, artist1, title1, artist2, title2)
	}
	if !c.IsCached(id) {
		t.Errorf("expectind IsCached == true for 2 songs")
	}
}
