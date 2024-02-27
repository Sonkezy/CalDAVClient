package caldavclient

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/pkg/errors"
)

type CaldavClient struct {
	caldav.Client
	Login     string
	Token     string
	Url       string
	principal string
	homeset   string
}

func NewCaldavClient(login, token, url string, ctx context.Context) (*CaldavClient, error) {
	caldavClient, err := caldav.NewClient(webdav.HTTPClientWithBasicAuth(http.DefaultClient, login, token), url)
	if err != nil {
		return nil, err
	}
	principal, err := caldavClient.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	homeset, err := caldavClient.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, err
	}
	return &CaldavClient{Client: *caldavClient, Login: login, Token: token, Url: url, principal: principal, homeset: homeset}, nil
}

func (c *CaldavClient) GetCalendarsNames(ctx context.Context) []string {
	calendars, err := c.FindCalendars(ctx, c.homeset)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var calendarNames []string
	for _, calendar := range calendars {
		calendarNames = append(calendarNames, calendar.Name)
	}
	return calendarNames
}

func (c *CaldavClient) GetCalendarsPaths(ctx context.Context) []string {
	calendars, err := c.FindCalendars(ctx, c.homeset)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var calendarPaths []string
	for _, calendar := range calendars {
		calendarPaths = append(calendarPaths, calendar.Path)
	}
	return calendarPaths
}

func (c *CaldavClient) GetCalendars(ctx context.Context, name string) ([]caldav.CalendarObject /*[]*ical.Component*/, error) {
	/*calendars, err := c.FindCalendars(ctx, c.homeset)
	if err != nil {
		return nil, err
	}
		var calendarExist bool = false
	for _, calendar := range calendars {
		if calendar.Name == name {
			calendarExist = true
		}
	}
	if !calendarExist {
		return nil, errors.New(fmt.Sprintf("Calendar %s not exist", name))
	}
	return nil, nil*/
	events, err := c.loadTodayEvents(ctx)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (c *CaldavClient) queryCalendarEventsByTimeRange(ctx context.Context,
	calendarPath string,
	start time.Time,
	end time.Time) ([]caldav.CalendarObject, error) {

	query := &caldav.CalendarQuery{
		CompFilter: caldav.CompFilter{
			Name: "VCALENDAR",
			Comps: []caldav.CompFilter{{
				Name:  "VEVENT",
				Start: start.UTC(),
				End:   end.UTC(),
			}},
		},
	}

	return c.QueryCalendar(ctx, calendarPath, query)
}

func (c *CaldavClient) GetTodayDateTimes() (time.Time, time.Time) {
	timeLocation, err := time.LoadLocation("Local")
	if err != nil {
		log.Println(err)
	}
	now := time.Now().In(timeLocation)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
	return start, end
}
func (c *CaldavClient) loadTodayEvents(ctx context.Context) ([]caldav.CalendarObject, error) {
	start, end := c.GetTodayDateTimes()
	var events []caldav.CalendarObject
	paths := c.GetCalendarsPaths(ctx)
	for _, path := range paths {
		tmp_events, err := c.LoadEvents(ctx, path, start, end)
		if err != nil {
			return nil, err
		}
		events = append(events, tmp_events...)
	}
	return events, nil
}

func (c *CaldavClient) LoadEvents(ctx context.Context, calendarPath string, start time.Time, end time.Time) ([]caldav.CalendarObject, error) {
	var events []caldav.CalendarObject
	calendarObjects, err := c.queryCalendarEventsByTimeRange(ctx, calendarPath, start, end)
	if calendarObjects == nil {
		log.Print("Can't get events for calendar "+calendarPath, " ", c.Login, " ", err)
		return events, errors.New("Can't get events from calendar")
	}
	return events, nil
}
