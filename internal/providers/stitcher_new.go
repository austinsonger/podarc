package providers

import (
	"encoding/json"
	"fmt"
	"github.com/sa7mon/podarc/internal/interfaces"
	"github.com/sa7mon/podarc/internal/utils"
	"log"
	"net/http"
	"strconv"
	"time"
)

type latestEpisodesResponse struct {
	Data struct {
		Shows []struct {
			ID              int         `json:"id"`
			ClassicID       interface{} `json:"classic_id"`
			Title           string      `json:"title"`
			Description     string      `json:"description"`
			HTMLDescription string      `json:"html_description"`
			EpisodeCount    int         `json:"episode_count"`
			DateCreated     int         `json:"date_created"`
			DatePublished   int         `json:"date_published"`
			ColorPrimary    string      `json:"color_primary"`
			ImageThumbnail  string      `json:"image_thumbnail"`
			ImageSmall      string      `json:"image_small"`
			ImageLarge      string      `json:"image_large"`
			ImageBaseURL    string      `json:"image_base_url"`
			DefaultSeasonID interface{} `json:"default_season_id"`
			DefaultSort     interface{} `json:"default_sort"`
			Link            interface{} `json:"link"`
			StitcherLink    string      `json:"stitcher_link"`
			SocialFacebook  interface{} `json:"social_facebook"`
			SocialTwitter   interface{} `json:"social_twitter"`
			SocialInstagram interface{} `json:"social_instagram"`
			Publisher       interface{} `json:"publisher"`
			IsPublished     bool        `json:"is_published"`
			IsPublic        bool        `json:"is_public"`
			Cadence         interface{} `json:"cadence"`
			Seasons         []struct {
				SeasonID int    `json:"season_id"`
				Name     string `json:"name"`
			} `json:"seasons"`
			Categories        []interface{} `json:"categories"`
			PrimaryCategoryID interface{}   `json:"primary_category_id"`
			Years             []int         `json:"years"`
			Restricted        []string      `json:"restricted"`
			Slug              string        `json:"slug"`
			Tags              []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				DisplayName string `json:"display_name"`
				Type        int    `json:"type"`
			} `json:"tags"`
		} `json:"shows"`
		Episodes []struct {
			ID                 int         `json:"id"`
			ShowID             int         `json:"show_id"`
			ClassicID          int         `json:"classic_id"`
			Title              string      `json:"title"`
			Description        string      `json:"description"`
			HTMLDescription    string      `json:"html_description"`
			Link               string      `json:"link"`
			StitcherLink       interface{} `json:"stitcher_link"`
			IsPublished        bool        `json:"is_published"`
			SeasonID           int         `json:"season_id"`
			Season             string      `json:"season"`
			AudioURL           string      `json:"audio_url"`
			AudioURLRestricted interface{} `json:"audio_url_restricted"`
			DateUpdated        int         `json:"date_updated"`
			DateCreated        int         `json:"date_created"`
			DatePublished      int64       `json:"date_published"`
			Duration           int         `json:"duration"`
			DurationRestricted interface{} `json:"duration_restricted"`
			Restriction        int         `json:"restriction"`
			GUID               string      `json:"guid"`
			Slug               string      `json:"slug"`
		} `json:"episodes"`
	} `json:"data"`
	Orchestration struct {
		StartIndex int `json:"start_index"`
		PageSize   int `json:"page_size"`
		TotalCount int `json:"total_count"`
	} `json:"orchestration"`
	Errors []interface{} `json:"errors"`
}

type StitcherNewPodcast struct {
	Name string
	Feed string
	ShowDescription string
	Episodes []StitcherNewEpisode
}

type StitcherNewEpisode struct {
	Id          string
	Image       string
	Published   time.Time
	Title       string
	Description string
	URL         string
}

func (s StitcherNewPodcast) NumEpisodes() int {
	return len(s.Episodes)
}

func (s StitcherNewPodcast) GetEpisodes() []interfaces.PodcastEpisode {
	// TODO: Might be more efficient to store these values rather than do a for loop every time the getter is called
	// Golang doesn't allow you to directly return a slice of a type as a slice of an interface
	// https://golang.org/doc/faq#convert_slice_of_interface
	intEpisodes := make([]interfaces.PodcastEpisode, len(s.Episodes))
	for i, elem := range s.Episodes {
		intEpisodes[i] = elem
	}
	return intEpisodes
}

func (s StitcherNewPodcast) GetTitle() string {
	return s.Name
}

func (s StitcherNewPodcast) GetDescription() string {
	return s.ShowDescription
}

func (s StitcherNewPodcast) GetPublisher() string {
	return "Stitcher"
}

func (e StitcherNewEpisode) GetTitle() string {
	return e.Title
}

func (e StitcherNewEpisode) GetDescription() string {
	return e.Description
}

func (e StitcherNewEpisode) GetURL() string {
	return e.URL
}

func (e StitcherNewEpisode) GetPublishedDate() string {
	return e.Published.String()
}

func (e StitcherNewEpisode) GetImageURL() string {
	return e.Image
}

func (e StitcherNewEpisode) GetParsedPublishedDate() (time.Time, error) {
	return e.Published, nil
}

func (e StitcherNewEpisode) ToString() string {
	return fmt.Sprintf("Title: %s | Description: %s | Url: %s | PublishedDate: " +
		"%s | ImageUrl: %s", e.GetTitle(), e.GetDescription(), e.GetURL(), e.GetPublishedDate(),
		e.GetImageURL())
}

func parseEpisodesFromResponse(response latestEpisodesResponse) []StitcherNewEpisode {
	var parsedEpisodes []StitcherNewEpisode
	for _, respEpisode := range response.Data.Episodes {
		var newEpisode StitcherNewEpisode
		newEpisode.Id = strconv.Itoa(respEpisode.ID)
		newEpisode.Image = response.Data.Shows[0].ImageBaseURL
		newEpisode.Published = time.Unix(respEpisode.DatePublished, 0)
		newEpisode.Title = respEpisode.Title
		newEpisode.Description = respEpisode.Description

		// Stitcher Premium-only episodes have the AudioURLRestricted field set. Otherwise, use AudioURL
		audioURLRestricted := fmt.Sprintf("%v", respEpisode.AudioURLRestricted)
		if audioURLRestricted != "" && audioURLRestricted != "null"{
			newEpisode.URL = audioURLRestricted
		} else {
			newEpisode.URL = respEpisode.AudioURL
		}

		parsedEpisodes = append(parsedEpisodes, newEpisode)
	}
	return parsedEpisodes
}

func GetStitcherNewPodcastFeed(slug string, creds string) *StitcherNewPodcast {
	/*
		The Stitcher API will return a practically unlimited number of episodes in a single page.
		This method will fetch up to 10,000 episodes on one page. If more than 10,000 are returned,
		we log.Fatal() and exit instead of missing episodes.
	*/

	valid, reason := utils.IsStitcherTokenValid(creds)
	if !valid {
		log.Fatal("Bad Stitcher token: " + reason)
	}
	stitcherPod := StitcherNewPodcast{}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", fmt.Sprintf(
		"https://api.prod.stitcher.com/show/%s/latestEpisodes?count=10000&page=0", slug), nil)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if resp.StatusCode != 200 {
		log.Fatal("Bad status code while getting podcast - " + resp.Status)
	}

	firstPageResponse := &latestEpisodesResponse{}

	jsonDecoder := json.NewDecoder(resp.Body)
	err = jsonDecoder.Decode(firstPageResponse)
	if err != nil {
		log.Fatal(err)
	}

	// The API doesn't currently have a page size limit. Fail here if that ever changes and we'll do proper paging.
	if firstPageResponse.Orchestration.TotalCount > firstPageResponse.Orchestration.PageSize {
		log.Fatal("Show has more than 1 page of episodes")
	}

	// Set podcast description, feed URL, and episodes from the first page
	stitcherPod.ShowDescription = firstPageResponse.Data.Shows[0].Description
	stitcherPod.Feed = firstPageResponse.Data.Shows[0].StitcherLink
	stitcherPod.Name = firstPageResponse.Data.Shows[0].Title

	firstPageEpisodes := parseEpisodesFromResponse(*firstPageResponse)
	stitcherPod.Episodes = firstPageEpisodes

	return &stitcherPod
}