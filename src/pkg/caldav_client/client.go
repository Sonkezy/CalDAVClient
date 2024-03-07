package caldavclient

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/emersion/go-ical"
	"github.com/lugamuga/go-webdav"
	"github.com/lugamuga/go-webdav/caldav"
	"github.com/pkg/errors"
)

type CaldavClient struct {
	caldav.Client
	Login     string
	Token     string
	URL       string
	principal string
	homeset   string
}

type EventClient struct {
	Calendar string
	Name     string
	Start    time.Time
	End      time.Time
	Location string
	path     string
}

func NewCaldavClient(login, token, url string) (*CaldavClient, error) {
	caldavClient, err := caldav.NewClient(webdav.HTTPClientWithBasicAuth(http.DefaultClient, login, token), url)
	if err != nil {
		return nil, err
	}
	principal, err := caldavClient.FindCurrentUserPrincipal()
	if err != nil {
		return nil, err
	}
	homeset, err := caldavClient.FindCalendarHomeSet(principal)
	if err != nil {
		return nil, err
	}
	return &CaldavClient{Client: *caldavClient, Login: login, Token: token, URL: url, principal: principal, homeset: homeset}, nil
}

func (c *CaldavClient) GetCalendarsNames() []string {
	calendars, err := c.FindCalendars(c.homeset)
	if err != nil {
		log.Println(err)
		return nil
	}
	var calendarNames []string
	for _, calendar := range calendars {
		calendarNames = append(calendarNames, calendar.Name)
	}
	return calendarNames
}

func (c *CaldavClient) GetCalendarsPaths() []string {
	calendars, err := c.FindCalendars(c.homeset)
	if err != nil {
		log.Println(err)
		return nil
	}
	var calendarPaths []string
	for _, calendar := range calendars {
		calendarPaths = append(calendarPaths, calendar.Path)
	}
	return calendarPaths
}

func (c *CaldavClient) GetCalendars(ctx context.Context, name string) ([]EventClient /*[]caldav.CalendarObject  []*ical.Component*/, error) {
	events, err := c.loadTodayEvents(ctx)
	if err != nil {
		return nil, err
	}
	eventsList, err := c.ParseEvents(events)
	if err != nil {
		return nil, err
	}
	return eventsList, nil
}

func (c *CaldavClient) queryCalendarEventsByTimeRange(
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

	return c.QueryCalendar(calendarPath, query)
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
	paths := c.GetCalendarsPaths()
	for _, path := range paths {
		tmpEvents, err := c.LoadEvents(ctx, path, start, end)
		if err != nil {
			return nil, err
		}
		events = append(events, tmpEvents...)
	}
	return events, nil
}

func (c *CaldavClient) LoadEvents(ctx context.Context, calendarPath string, start time.Time, end time.Time) ([]caldav.CalendarObject, error) {
	calendarObjects, err := c.queryCalendarEventsByTimeRange(calendarPath, start, end)
	if calendarObjects == nil {
		log.Print("Can't get events for calendar "+calendarPath, " ", c.Login, " ", err)
		return calendarObjects, errors.New("Can't get events from calendar")
	}
	return calendarObjects, nil
}

func (c *CaldavClient) ParseEvents(objects []caldav.CalendarObject) ([]EventClient, error) {
	var eventsList []EventClient
	timeLocation, err := time.LoadLocation("Local")
	if err != nil {
		log.Println(err)
	}
	for _, object := range objects {
		startTime, err := time.ParseInLocation("20060102T150405Z", (*object.Data).Component.Children[0].Props["DTSTART"][0].Value, timeLocation)
		if err != nil {
			return nil, err
		}
		endTime, err := time.ParseInLocation("20060102T150405Z", (*object.Data).Component.Children[0].Props["DTEND"][0].Value, timeLocation)
		if err != nil {
			return nil, err
		}
		var location string
		if (*object.Data).Component.Children[0].Props["LOCATION"] != nil {
			location = (*object.Data).Component.Children[0].Props["LOCATION"][0].Value
		}
		event := &EventClient{
			Name:     (*object.Data).Component.Children[0].Props["SUMMARY"][0].Value,
			Location: location,
			Start:    startTime,
			End:      endTime,
			path:     object.Path,
		}
		eventsList = append(eventsList, *event)
	}
	return eventsList, nil
}

func (c *CaldavClient) OutputEvents(eventsList []EventClient) {
	for n, event := range eventsList {
		fmt.Printf("%d) Name: %s\n", n+1, event.Name)
		fmt.Printf("----Calendar: %s\n", event.Calendar)
		fmt.Printf("----Location: %s\n", event.Location)
		fmt.Printf("----Start time: %s\n", event.Start.Format(time.UnixDate))
		fmt.Printf("----End time:   %s\n", event.End.Format(time.UnixDate))
		fmt.Printf("----Path:   %s\n", event.path)
	}
}

func (c *CaldavClient) CreateEvent() {
	newEvent := &EventClient{}
	fmt.Println("Type event name: ")
	fmt.Scan(&newEvent.Name)
	fmt.Println("Type location name: ")
	fmt.Scan(&newEvent.Location)
	fmt.Println("Type event start time (example: 2000-12-31 12:00:00): ")
	reader := bufio.NewReader(os.Stdin)
	startTime, _ := reader.ReadString('\n')
	startTime = startTime[:len(startTime)-1]
	var err error
	newEvent.Start, err = time.Parse("2006-01-02 03:04:05", startTime)
	if err != nil {
		fmt.Println("Unexpected input ", err)
		return
	}
	fmt.Println("Type event end time (format ): ")
	endTime, _ := reader.ReadString('\n')
	if len(endTime) > 19 {
		endTime = endTime[:19]
	}
	newEvent.End, err = time.Parse("2006-01-02 03:04:05", endTime)
	if err != nil {
		fmt.Println("Unexpected input ", err)
		return
	}
	if startTime > endTime {
		fmt.Println("Start time later then end time")
		return
	}
	c.PutEvent(*newEvent)
}

func (c *CaldavClient) PutEvent(event EventClient) {
	calendarsPath := c.GetCalendarsPaths()
	calendarsNames := c.GetCalendarsNames()
	fmt.Println("Select calendar:")
	for n, name := range calendarsNames {
		fmt.Printf("%d) %s\n", n+1, name)
	}
	var calendarNumber int
	fmt.Scan(&calendarNumber)
	calendarNumber--
	if calendarNumber < 0 || calendarNumber >= len(calendarsPath) {
		fmt.Println("Wrong number")
		return
	}
	newEvent := ical.NewCalendar()
	newEvent.Component.Props.Add(&ical.Prop{
		Name:  "PRODID",
		Value: "-//CalDAV //CalDAV Calendar//EN",
	})
	newEvent.Component.Props.Add(&ical.Prop{
		Name:  "VERSION",
		Value: "2.0",
	})
	newComponent := ical.NewComponent("VEVENT")
	newComponent.Props.Add(&ical.Prop{
		Name:  "SUMMARY",
		Value: event.Name,
	})
	newComponent.Props.Add(&ical.Prop{
		Name:  "LOCATION",
		Value: event.Location,
	})
	newComponent.Props.Add(&ical.Prop{
		Name:  "DTSTART",
		Value: event.Start.Format("20060102T150405"),
	})
	newComponent.Props.Add(&ical.Prop{
		Name:  "DTEND",
		Value: event.End.Format("20060102T150405"),
	})
	newComponent.Props.Add(&ical.Prop{
		Name:  "DTSTAMP",
		Value: event.Start.Format("20060102T150405"),
	})
	newComponent.Props.Add(&ical.Prop{
		Name:  "UID",
		Value: "testN1",
	})
	newEvent.Component.Children = append(newEvent.Component.Children, newComponent)
	_, err := c.PutCalendarObject(c.homeset, newEvent)
	if err != nil {
		log.Println(err)
	}
}
