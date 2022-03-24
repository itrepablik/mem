package mem

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Schedule Types
const (
	FREQUENTLY = iota
	DAILY
	WEEKLY
	MONTHLY
)

// Frequently interval options, avoid 0 as it is the default
const (
	EVERY_SECOND = iota + 1
	EVERY_MINUTE
	EVERY_HOUR
)

// Day name options, avoid 0 as it is the default
const (
	SUNDAY = iota + 1
	MONDAY
	TUESDAY
	WEDNESDAY
	THURSDAY
	FRIDAY
	SATURDAY
)

// Common cleaner config options
const (
	FREQUENTLY_SCHEDULE_TYPE = "frequently"
	DAILY_SCHEDULE_TYPE      = "daily"
	WEEKLY_SCHEDULE_TYPE     = "weekly"
	MONTHLY_SCHEDULE_TYPE    = "monthly"
	FREQUENTLY_EVERY_SECOND  = "EVERY_SECOND"
	FREQUENTLY_EVERY_MINUTE  = "EVERY_MINUTE"
	FREQUENTLY_EVERY_HOUR    = "EVERY_HOUR"
	DT_FORMAT                = "2006-01-02 15:04:05"
	DEFAULT_START_TIME       = "00:00:00"
)

// CleanerSchedule is a struct that holds the data for the cleaner schedule
type CleanerSchedule struct {
	ScheduleType  int    // FREQUENTLY, DAILY, WEEKLY, MONTHLY
	Interval      int    // interval for the schedule
	IntervalValue int    // interval value for the schedule
	StartTime     string // input time for the schedule using 24 hour format e.g 23:00
}

// Cleaner is a struct that holds the data for the cleaner
type Cleaner struct {
	TaskName string // name of the task
	Schedule *CleanerSchedule
	LastRun  int64         // unix timestamp of when the last run was
	NextRun  int64         // unix timestamp of when the next run will be
	Remarks  string        // last run remarks
	mu       *sync.RWMutex // read-write mutex, multiple readers, single writer
}

// CleanerOption is a cleaner option interface
type CleanerOption interface {
	IntervalOpt() int
	IntervalValueOpt() int
	StartTimeOpt() string
	SetInterval(int)
	SetIntervalValue(int)
	SetStartTime(string)
	Error() error
}

// IntervalOpt returns the interval for the cleaner
func (c *CleanerSchedule) IntervalOpt() int {
	return c.Interval
}

// SetInterval sets the interval for the cleaner
func (c *CleanerSchedule) SetInterval(interval int) {
	c.Interval = interval
}

// IntervalValueOpt returns the interval value for the cleaner
func (c *CleanerSchedule) IntervalValueOpt() int {
	return c.IntervalValue
}

// SetIntervalValue sets the interval value for the cleaner
func (c *CleanerSchedule) SetIntervalValue(intervalValue int) {
	c.IntervalValue = intervalValue
}

// StartTimeOpt returns the start time for the cleaner
func (c *CleanerSchedule) StartTimeOpt() string {
	return c.StartTime
}

// SetStartTime sets the start time for the cleaner
func (c *CleanerSchedule) SetStartTime(startTime string) {
	c.StartTime = startTime
}

// Error returns the error for the cleaner
func (cs *CleanerSchedule) Error() error {
	// Get the schedule type name
	schedTypeName := getSchedTypeName(cs)
	if len(strings.TrimSpace(schedTypeName)) == 0 {
		schedTypeName = "unknown: " + fmt.Sprintf("%d", cs.ScheduleType)
	}

	// Validates the schedule type
	isValidScheduleType := false
	switch cs.ScheduleType {
	case FREQUENTLY, DAILY, WEEKLY, MONTHLY:
		isValidScheduleType = true
	}
	if !isValidScheduleType {
		return fmt.Errorf("invalid schedule type: %s", schedTypeName)
	}

	isValidInterval := false
	switch cs.ScheduleType {
	case FREQUENTLY:
		switch cs.Interval {
		case EVERY_SECOND, EVERY_MINUTE, EVERY_HOUR:
			// Start time is not required for the frequently schedule
			switch len(strings.TrimSpace(cs.StartTime)) {
			case 0:
				isValidInterval = true
			default:
				return fmt.Errorf("start time input for the schedule type %s is not allowed", schedTypeName)
			}
		default:
			return fmt.Errorf("invalid interval: %d, options are: %s, %s, %s", cs.Interval,
				FREQUENTLY_EVERY_SECOND, FREQUENTLY_EVERY_MINUTE, FREQUENTLY_EVERY_HOUR)
		}

	case DAILY:
		// It will start in the next day at the start time of the day
		// If no start time is provided, then it's an invalid interval
		switch len(strings.TrimSpace(cs.StartTime)) {
		case 0:
			return fmt.Errorf("invalid start time: %s", cs.StartTime)
		default:
			isValidInterval = true
		}

	case WEEKLY:
		switch cs.Interval {
		case SUNDAY, MONDAY, TUESDAY, WEDNESDAY, THURSDAY, FRIDAY, SATURDAY:
			// If no start time is provided, then it's an invalid interval
			switch len(strings.TrimSpace(cs.StartTime)) {
			case 0:
				return fmt.Errorf("invalid start time: %s", cs.StartTime)
			default:
				isValidInterval = true
			}
		}

	case MONTHLY:
		// If no start time is provided, then it's an invalid interval
		switch len(strings.TrimSpace(cs.StartTime)) {
		case 0:
			return fmt.Errorf("invalid start time: %s", cs.StartTime)
		default:
			isValidInterval = true
		}

		// Day validation from 1-31 days only
		switch {
		case cs.Interval >= 1 && cs.Interval <= 31:
			isValidInterval = true
		default:
			return fmt.Errorf("invalid interval: %d, options are: 1-31", cs.Interval)
		}
	}

	if !isValidInterval {
		return fmt.Errorf("invalid interval: %d for the schedule type: %s", cs.Interval, schedTypeName)
	}
	return nil
}

// WithInterval sets the interval for the cleaner
func WithInterval(interval int) CleanerOption {
	return &CleanerSchedule{Interval: interval}
}

// WithIntervalValue sets the interval value for the cleaner
func WithIntervalValue(interval, value int) CleanerOption {
	return &CleanerSchedule{
		Interval:      interval,
		IntervalValue: value,
	}
}

// WithStartTime sets the start time for the cleaner
func WithStartTime(startTime string) CleanerOption {
	return &CleanerSchedule{StartTime: startTime}
}

// WithWeekDay sets the weekday for the cleaner
func WithWeekDay(weekday int) CleanerOption {
	return &CleanerSchedule{Interval: weekday}
}

// WithDayOfMonth sets the day of month for the cleaner
func WithDayOfMonth(dayOfMonth int) CleanerOption {
	return &CleanerSchedule{Interval: dayOfMonth}
}

// NewCleaner creates a new cleaner
func NewCleaner(scheduleType int, opts ...CleanerOption) (*Cleaner, error) {
	c := &Cleaner{
		Schedule: &CleanerSchedule{
			ScheduleType: scheduleType,
		},
		mu: &sync.RWMutex{},
	}

	// Apply the options
	for _, opt := range opts {
		// Check if Interval option is set
		switch opt.IntervalOpt() {
		case 0:
			// No interval is set
		default:
			c.Schedule.SetInterval(opt.IntervalOpt())
		}

		// Check if IntervalValue option is set
		switch opt.IntervalValueOpt() {
		case 0:
			// No interval value is set
		default:
			c.Schedule.SetIntervalValue(opt.IntervalValueOpt())
		}

		// Check if StartTime option is set
		switch opt.StartTimeOpt() {
		case "":
			// No start time is set
		default:
			c.Schedule.SetStartTime(opt.StartTimeOpt())
		}
	}

	// Check for errors
	if err := c.Schedule.Error(); err != nil {
		return nil, err
	}
	return c, nil
}

// Command returns the command to run the cleaner
type Command interface {
	Run(h *Cache)
}

// ChannelTS is a channel timestamp
var ChannelTS = make(chan bool, 1)

// Run runs the cleaner
func (c *Cleaner) Run(h *Cache) {
	switch c.Schedule.ScheduleType {
	case FREQUENTLY:
		c.CleanFrequently()
	case DAILY:
		c.CleanDaily()
	case WEEKLY:
		c.CleanWeekly()
	case MONTHLY:
		c.CleanMonthly()
	}

	// Add the cleaner to the list
	c.AddCleaner()

	// Itereate over the list of cleaners and run them
	for {
		select {
		case <-ChannelTS:
			return
		case <-time.After(500 * time.Millisecond):
			go execRunner(c, h)
		}
	}
}

// execRunner is the runner for the exec command
func execRunner(c *Cleaner, h *Cache) {
	// Scan the list of cleaners and run them
	for _, e := range GetAllCleanerSchedules() {
		for _, s := range e {
			// Check if due for execution
			if s.NextRun == time.Now().Unix() {
				c.UpdateNextRun(s.TaskName) // Update the next run time for the task
				CleanExpired(h)
			}
		}
	}
}

// UpdateNextRun updates the next run time
func (c *Cleaner) UpdateNextRun(taskName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.Schedule.ScheduleType {
	case FREQUENTLY:
		// Check the interval value
		if c.Schedule.IntervalValue <= 0 {
			c.Schedule.IntervalValue = 1
		}

		switch c.Schedule.Interval {
		case EVERY_SECOND:
			c.NextRun = time.Now().Add(time.Second * time.Duration(c.Schedule.IntervalValue)).Unix()

		case EVERY_MINUTE:
			c.NextRun = time.Now().Add(time.Minute * time.Duration(c.Schedule.IntervalValue)).Unix()

		case EVERY_HOUR:
			c.NextRun = time.Now().Add(time.Hour * time.Duration(c.Schedule.IntervalValue)).Unix()
		}

	case DAILY:
		// Get the start time hour and minute
		startTimeHour, startTimeMinute := GetTime(c.Schedule.StartTime)

		// Start the cleaner for the following day with start time
		c.NextRun = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()+1, startTimeHour, startTimeMinute, 0, 0, time.Local).Unix()

	case WEEKLY:
		// Get the start time hour and minute
		startTimeHour, startTimeMinute := GetTime(c.Schedule.StartTime)
		dayOfWeek := c.Schedule.Interval - 1

		// Find the next dayOfWeek in the future from the current day
		nextDayOfWeek := time.Now().AddDate(0, 0, (int(dayOfWeek) + int(time.Now().Weekday()) - 1))
		c.NextRun = time.Date(nextDayOfWeek.Year(), nextDayOfWeek.Month(), nextDayOfWeek.Day(), startTimeHour, startTimeMinute, 0, 0, time.Local).Unix()

	case MONTHLY:
		// Get the start time hour and minute
		startTimeHour, startTimeMinute := GetTime(c.Schedule.StartTime)

		// Find the next dayOfMonth in the future from the current day
		c.NextRun = time.Date(time.Now().Year(), time.Now().Month()+1, c.Schedule.Interval, startTimeHour, startTimeMinute, 0, 0, time.Local).Unix()
	}

	c.Remarks = fmt.Sprintf("%s ran successfully on %s", taskName, time.Now().Format(DT_FORMAT))

	// Update cleaner TS with the new next run time
	UpdateCleaner(c, taskName)
}

// CleanerScheduler is the collection of cleaners that are scheduled to run
type CleanerScheduler struct {
	CleanerList map[string][]Cleaner
	mu          sync.RWMutex
}

// TS initialize the 'CleanerScheduler' struct with an empty values
var TS = CleanerScheduler{CleanerList: make(map[string][]Cleaner)}

// AddCleaner to the list of cleaners to run
func (c *Cleaner) AddCleaner() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Generate secure random string using rand package
	b := make([]byte, 16)
	rand.Read(b)
	key := fmt.Sprintf("%x", b)

	// Check if the cleaner is already added by using GetCleanerSchedule method
	if _, err := TS.GetCleanerSchedule(key); err == nil {
		return // If the cleaner is already added, then just return
	}
	c.TaskName = key

	// Add the cleaner to the list
	TS.mu.Lock()
	TS.CleanerList[key] = append(TS.CleanerList[key], *c)
	TS.mu.Unlock()
}

// UpdateCleaner update the cleaner with the new one
func UpdateCleaner(c *Cleaner, taskName string) {
	TS.mu.Lock()
	TS.CleanerList[taskName] = []Cleaner{*c}
	TS.mu.Unlock()
}

// GetCleanerSchedule returns the cleaner by the key
func (s *CleanerScheduler) GetCleanerSchedule(key string) (Cleaner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the key is present
	if _, ok := s.CleanerList[key]; !ok {
		return Cleaner{}, fmt.Errorf("cleaner not found for key: %s", key)
	}
	return s.CleanerList[key][0], nil
}

// GetAllCleanerSchedules returns the list of cleaners
func GetAllCleanerSchedules() map[string][]Cleaner {
	TS.mu.Lock()
	defer TS.mu.Unlock()
	return TS.CleanerList
}

// CleanFrequently runs the cleaner frequently
func (c *Cleaner) CleanFrequently() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check the interval value
	if c.Schedule.IntervalValue <= 0 {
		c.Schedule.IntervalValue = 1
	}

	switch c.Schedule.Interval {
	case EVERY_SECOND:
		c.NextRun = time.Now().Add(time.Second * time.Duration(c.Schedule.IntervalValue)).Unix()
	case EVERY_MINUTE:
		c.NextRun = time.Now().Add(time.Minute * time.Duration(c.Schedule.IntervalValue)).Unix()
	case EVERY_HOUR:
		c.NextRun = time.Now().Add(time.Hour * time.Duration(c.Schedule.IntervalValue)).Unix()
	}
}

// CleanDaily runs the cleaner daily
func (c *Cleaner) CleanDaily() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the start time hour and minute
	startTimeHour, startTimeMinute := GetTime(c.Schedule.StartTime)

	// Start the cleaner for the following day with start time
	c.NextRun = time.Date(time.Now().Year(), time.Now().Month(),
		time.Now().Day()+1, startTimeHour, startTimeMinute, 0, 0, time.Local).Unix()
}

// CleanWeekly runs the cleaner weekly
func (c *Cleaner) CleanWeekly() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the start time hour and minute
	startTimeHour, startTimeMinute := GetTime(c.Schedule.StartTime)
	dayOfWeek := c.Schedule.Interval - 1

	// Find the next dayOfWeek in the future from the current day
	nextDayOfWeek := time.Now().AddDate(0, 0, (int(dayOfWeek) + int(time.Now().Weekday()) - 1))
	c.NextRun = time.Date(nextDayOfWeek.Year(), nextDayOfWeek.Month(),
		nextDayOfWeek.Day(), startTimeHour, startTimeMinute, 0, 0, time.Local).Unix()
}

// CleanMonthly runs the cleaner monthly
func (c *Cleaner) CleanMonthly() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get the start time hour and minute
	startTimeHour, startTimeMinute := GetTime(c.Schedule.StartTime)

	// Find the next dayOfMonth in the future from the current day
	c.NextRun = time.Date(time.Now().Year(), time.Now().Month()+1,
		c.Schedule.Interval, startTimeHour, startTimeMinute, 0, 0, time.Local).Unix()
}

// GetTime returns the hour and minute of the time
func GetTime(startTime string) (int, int) {
	// Get the start time hour and minute
	startTimeHour, err := strconv.Atoi(strings.Split(startTime, ":")[0])
	if err != nil {
		return 0, 0
	}

	startTimeMinute, err := strconv.Atoi(strings.Split(startTime, ":")[1])
	if err != nil {
		return 0, 0
	}
	return startTimeHour, startTimeMinute
}

// getSchedTypeName returns the schedule type name
func getSchedTypeName(schedule *CleanerSchedule) string {
	switch schedule.ScheduleType {
	case FREQUENTLY:
		return FREQUENTLY_SCHEDULE_TYPE
	case DAILY:
		return DAILY_SCHEDULE_TYPE
	case WEEKLY:
		return WEEKLY_SCHEDULE_TYPE
	case MONTHLY:
		return MONTHLY_SCHEDULE_TYPE
	}
	return ""
}
