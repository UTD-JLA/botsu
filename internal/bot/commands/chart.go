package commands

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"math"
	"net/http"
	"net/url"
	"text/template"

	"github.com/UTD-JLA/botsu/internal/activities"
	"github.com/UTD-JLA/botsu/internal/bot"
	"github.com/UTD-JLA/botsu/internal/users"
	"github.com/UTD-JLA/botsu/pkg/discordutil"
	"github.com/bwmarrin/discordgo"
	"github.com/golang-module/carbon/v2"
)

var ChartCommandData = &discordgo.ApplicationCommand{
	Name:        "chart",
	Description: "View a chart of your activity",
	Options:     []*discordgo.ApplicationCommandOption{},
}

type ChartCommand struct {
	ar *activities.ActivityRepository
	ur *users.UserRepository
}

func NewChartCommand(ar *activities.ActivityRepository, ur *users.UserRepository) *ChartCommand {
	return &ChartCommand{ar: ar, ur: ur}
}

var quickChartURL = url.URL{
	Scheme: "https",
	Host:   "quickchart.io",
	Path:   "/chart/create",
}

//go:embed chart_body.json.tmpl
var bodyTemplateFile string
var bodyTemplate = template.Must(template.New("body").Parse(bodyTemplateFile))

type body struct {
	Values      string
	Labels      string
	Color       string
	Annotations string
}

type response struct {
	Success bool   `json:"success"`
	Url     string `json:"url"`
}

func colorAsHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}

func getQuickChartBody(xValues []string, yValues []float64) bytes.Buffer {
	buffer := bytes.Buffer{}

	values, _ := json.Marshal(yValues)
	labels, _ := json.Marshal(xValues)

	bodyTemplate.Execute(&buffer, body{
		Values:      string(values),
		Labels:      string(labels),
		Color:       fmt.Sprintf("\"%s\"", colorAsHex(discordutil.ColorSecondary)),
		Annotations: "[]",
	})

	comactBuffer := bytes.Buffer{}
	json.Compact(&comactBuffer, buffer.Bytes())

	return comactBuffer
}

func (c *ChartCommand) Handle(ctx *bot.InteractionContext) error {
	userID := discordutil.GetInteractionUser(ctx.Interaction()).ID
	user, err := c.ur.FindOrCreate(ctx.ResponseContext(), userID)

	if err != nil {
		return err
	}

	timezone := carbon.UTC

	if user.Timezone != nil {
		timezone = *user.Timezone
	}

	start := carbon.Now(timezone).SubDays(6).StartOfDay()
	end := carbon.Now(timezone).EndOfDay()

	dailyDurations, err := c.ar.GetTotalByUserIDGroupedByDay(
		ctx.ResponseContext(),
		user.ID,
		start.ToStdTime(),
		end.ToStdTime(),
	)

	if err != nil {
		return err
	}

	totalMinutes := 0.0
	highestMinutes := 0.0
	highestDay := "N/A"

	values := make([]float64, 0, dailyDurations.Len())

	for _, k := range dailyDurations.Keys() {
		v, _ := dailyDurations.Get(k)
		totalMinutes += v.Minutes()
		if v.Minutes() > highestMinutes {
			highestMinutes = v.Minutes()
			highestDay = k
		}
		values = append(values, v.Minutes())
	}

	avgMinutes := totalMinutes / float64(dailyDurations.Len())

	reqBody := getQuickChartBody(dailyDurations.Keys(), values)
	req := http.Request{
		Method: http.MethodPost,
		URL:    &quickChartURL,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(&reqBody),
	}

	resp, err := http.DefaultClient.Do(&req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var quickChartResponse response
	err = json.NewDecoder(resp.Body).Decode(&quickChartResponse)

	if err != nil {
		return err
	}

	if !quickChartResponse.Success {
		return errors.New("failed to generate chart")
	}

	return ctx.Respond(discordgo.InteractionResponseChannelMessageWithSource, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{
			discordutil.NewEmbedBuilder().
				SetTitle("Activity History").
				SetColor(discordutil.ColorPrimary).
				SetImage(quickChartResponse.Url).
				AddField("Total", fmt.Sprintf("%.0f minutes", math.Round(totalMinutes)), true).
				AddField("Average", fmt.Sprintf("%.0f minutes", math.Round(avgMinutes)), true).
				AddField("Highest", fmt.Sprintf("%.0f minutes (%s)", math.Round(highestMinutes), highestDay), true).
				Build(),
		},
	})
}
