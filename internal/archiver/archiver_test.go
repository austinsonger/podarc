package archiver

import (
	"fmt"
	"github.com/sa7mon/podarc/internal/interfaces"
	"github.com/sa7mon/podarc/internal/providers"
	"github.com/sa7mon/podarc/internal/utils"
	"github.com/sa7mon/podarc/test"
	"os"
	"testing"
)

func TestGetFileNameFromEpisodeURL(t *testing.T) {
	ep := providers.GenericEpisode{}
	ep.Enclosure.URL = "https://my.site/file.mp3"

	fileName, err := GetFileNameFromEpisodeURL(ep)
	if err != nil {
		t.Error(err)
	}

	test.AssertString(t, "fileName", "file.mp3", fileName)

	ep.Enclosure.URL = "https://my.site/file.mp3?asdf=1"
	test.AssertString(t, "fileName2", "file.mp3", fileName)

	ep2 := providers.GenericEpisode{}
	ep2.Enclosure.URL = "{}[]_=__++!@#$%A^&*()()()"
	fileName, err = GetFileNameFromEpisodeURL(ep2)
	if err == nil {
		t.Error("Bad URL didn't return an error")
	}
}

func TestGetEpisodeFileName(t *testing.T) {
	genericEpisode := providers.GenericEpisode{Title: "My Cool Episode", PubDate: "Mon, 02 Jan 2006 15:04:05 -0700"}

	name1 := GetEpisodeFileName("my_cool_episode.mp3", genericEpisode)
	test.AssertString(t, "GetEpisodeFileName_good", "2006-01-02_my-cool-episode.mp3", name1)

	name2 := GetEpisodeFileName("my_cool_episode.mp3?tracking_tag=asdf", genericEpisode)
	test.AssertString(t, "GetEpisodeFileName_good", "2006-01-02_my-cool-episode.mp3", name2)
}

func TestWriteID3TagsToFile(t *testing.T) {
	// Test setup
	err := utils.DownloadFile("test_id3_tagging.bin", "https://fastest.fish/lib/downloads/1KB.bin", nil,false)
	if err != nil {
		t.Error(err)
	}

	// Test
	testPod := providers.GenericPodcast{}
	testPod.Channel.Title = "My Cool Podcast"
	testEpisode := providers.GenericEpisode{
		Title: "My Test Episode",
		PubDate: "Mon, 02 Jan 2006 15:04:05 -0700",
	}

	err = WriteID3TagsToFile("test_id3_tagging.bin", testEpisode, testPod)
	if err != nil {
		t.Error(err)
	}

	// Test cleanup
	err = os.Remove("test_id3_tagging.bin")
	if err != nil {
		fmt.Println("Couldn't delete test file")
	}

	err = WriteID3TagsToFile("non-existent-file.lol", testEpisode, testPod)
	if err == nil {
		t.Error("writing tags to non-existent file didn't return an error")
	}
}

func TestArchivePodcast(t *testing.T) {
	testPod := providers.GenericPodcast{}
	testPod.Channel.Title = "My Cool Podcast"
	testPod.Episodes = []interfaces.PodcastEpisode{}

	testEpisode := providers.GenericEpisode{
		Title: "My Test Episode",
		PubDate: "Mon, 02 Jan 2006 15:04:05 -0700",
	}
	testEpisode.Enclosure.URL = "https://fastest.fish/lib/downloads/1KB.bin"
	testPod.Episodes = append(testPod.Episodes, testEpisode)

	err := ArchivePodcast(testPod, "./", false, true, utils.Credentials{})
	if err != nil {
		t.Error(err)
	}
	err = os.Remove("1KB.bin.tmp")
	if err != nil {
		fmt.Println("Couldn't delete test file: " + err.Error())
	}

}