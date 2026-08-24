// Package api implements the HTTP handlers for user-created stories. This file
// is the P3.3 card: the DB → camelCase story JSON adapter. StoryView converts a
// stories row plus its chapters into the exact legacy story JSON shape the
// frontend renderer consumes (the same shape as the embedded *-storymap.json
// files). It maps snake_case/int DB values to camelCase and omits empty media
// fields.
package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Defaults for legacy top-level fields that are not stored in the stories
// table. The DB schema keeps only the columns the API needs; the renderer
// expects the full legacy shape, so the adapter fills the rest with sensible
// defaults.
const (
	defaultStyle         = "https://tiles.openfreemap.org/styles/liberty"
	defaultTheme         = "dark"
	defaultInsetPosition = "bottom-left"
	defaultStartSlide    = "none"
	defaultEndSlide      = "none"
)

// defaultGlobalView is the fallback camera used when a story has no global_view.
var defaultGlobalView = json.RawMessage(`{"center":[0,20],"zoom":0.6,"pitch":0,"bearing":0}`)

// StoryViewDoc is the camelCase legacy story JSON shape (top level).
type StoryViewDoc struct {
	Title         string          `json:"title"`
	Subtitle      string          `json:"subtitle"`
	Byline        string          `json:"byline"`
	Footer        string          `json:"footer"`
	Theme         string          `json:"theme"`
	Style         string          `json:"style"`
	InsetWidth    int             `json:"insetWidth"`
	InsetHeight   int             `json:"insetHeight"`
	InsetPosition string          `json:"insetPosition"`
	GlobalView    json.RawMessage `json:"globalView"`
	StartSlide    string          `json:"startSlide"`
	EndSlide      string          `json:"endSlide"`
	Chapters      []ChapterView   `json:"chapters"`
}

// ChapterView is the camelCase legacy chapter shape. Media fields (image/video/
// audio) are omitted when empty, matching the legacy renderer's expectations.
type ChapterView struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	Alignment       string          `json:"alignment"`
	Hidden          bool            `json:"hidden"`
	Location        *Location       `json:"location,omitempty"`
	MapAnimation    string          `json:"mapAnimation"`
	RotateAnimation bool            `json:"rotateAnimation"`
	OnChapterEnter  json.RawMessage `json:"onChapterEnter,omitempty"`
	OnChapterExit   json.RawMessage `json:"onChapterExit,omitempty"`
	Source          json.RawMessage `json:"source,omitempty"`
	Image           string          `json:"image,omitempty"`
	Video           string          `json:"video,omitempty"`
	Audio           string          `json:"audio,omitempty"`
	AutoPlayAudio   bool            `json:"autoPlayAudio"`
}

// StoryView converts a stories row and its chapters into the exact legacy
// story JSON shape the renderer consumes. It maps snake_case/int DB values to
// camelCase and omits empty media fields. This is the P3.3 HANDOFF.
func StoryView(s Story, chapters []Chapter) any {
	doc := StoryViewDoc{
		Title:         s.Title,
		Subtitle:      s.Subtitle,
		Byline:        s.Byline,
		Footer:        "",
		Theme:         defaultTheme,
		Style:         defaultStyle,
		InsetWidth:    0,
		InsetHeight:   0,
		InsetPosition: defaultInsetPosition,
		GlobalView:    defaultGlobalView,
		StartSlide:    defaultStartSlide,
		EndSlide:      defaultEndSlide,
		Chapters:      make([]ChapterView, 0, len(chapters)),
	}
	for _, c := range chapters {
		doc.Chapters = append(doc.Chapters, chapterView(c))
	}
	return doc
}

// chapterView maps a DB chapter row to the legacy camelCase chapter shape.
func chapterView(c Chapter) ChapterView {
	v := ChapterView{
		ID:              strconv.FormatInt(c.ID, 10),
		Title:           c.Title,
		Description:     c.DescriptionMD,
		Alignment:       c.Alignment,
		Hidden:          c.Hidden,
		MapAnimation:    c.MapAnimation,
		RotateAnimation: c.RotateAnimation,
		AutoPlayAudio:   false,
	}
	if c.Location != nil {
		v.Location = c.Location
	}
	if len(c.OnChapterEnter) > 0 {
		v.OnChapterEnter = c.OnChapterEnter
	}
	if len(c.OnChapterExit) > 0 {
		v.OnChapterExit = c.OnChapterExit
	}
	if c.Source != "" {
		v.Source = json.RawMessage(c.Source)
	}
	switch c.MediaType {
	case "image":
		v.Image = mediaURL(c)
	case "video":
		v.Video = mediaURL(c)
	case "audio":
		v.Audio = mediaURL(c)
	}
	return v
}

// mediaURL resolves a chapter's media to the URL the renderer loads directly.
// External refs use the stored URL; local refs point at the media serve path
// (P4.x serves /media/:aid). Empty refs yield "" (omitted by omitempty).
func mediaURL(c Chapter) string {
	switch c.MediaRefType {
	case "external":
		return c.MediaExternalURL
	case "local":
		if c.MediaAssetID != nil {
			return fmt.Sprintf("/media/%d", *c.MediaAssetID)
		}
	}
	return ""
}
