package domain

type HeroContent struct {
	Headline        string `json:"headline"`
	Subheadline     string `json:"subheadline"`
	CTAText         string `json:"cta_text"`
	CTAURL          string `json:"cta_url"`
	BackgroundImage string `json:"background_image"`
}

type FeaturesContent struct {
	SectionTitle       string        `json:"section_title"`
	SectionDescription string        `json:"section_description"`
	Items              []FeatureItem `json:"items"`
}

type FeatureItem struct {
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type EpisodesShowcaseContent struct {
	SectionTitle       string `json:"section_title"`
	SectionDescription string `json:"section_description"`
	MaxEpisodes        int    `json:"max_episodes"`
	DisplayMode        string `json:"display_mode"`
}

type TestimonialsContent struct {
	SectionTitle       string            `json:"section_title"`
	SectionDescription string            `json:"section_description"`
	Items              []TestimonialItem `json:"items"`
}

type TestimonialItem struct {
	Quote     string `json:"quote"`
	Author    string `json:"author"`
	Role      string `json:"role"`
	AvatarURL string `json:"avatar_url"`
}

type CTAContent struct {
	Headline    string `json:"headline"`
	Description string `json:"description"`
	ButtonText  string `json:"button_text"`
	ButtonURL   string `json:"button_url"`
}

type FooterContent struct {
	Copyright   string       `json:"copyright"`
	Links       []FooterLink `json:"links"`
	SocialLinks []SocialLink `json:"social_links"`
}

type FooterLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
}

type ProseContent struct {
	Body string `json:"body"`
}
