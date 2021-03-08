package main

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestFillInGaps(t *testing.T) {
	gapCheck := func(fixed []Datapoint) error {
		gap := fixed[1].DateTime.Unix() - fixed[0].DateTime.Unix()
		for i := 2; i < len(fixed); i++ {
			thisGap := fixed[i].DateTime.Unix() - fixed[i-1].DateTime.Unix()
			if thisGap != gap {
				return fmt.Errorf("unequal gaps: (i=%d) %d - %d = %d; (i=%d) %d - %d = %d;", i, fixed[i].DateTime.Unix(), fixed[i-1].DateTime.Unix(), thisGap, i-1, fixed[i-1].DateTime.Unix(), fixed[i-2].DateTime.Unix(), gap)
			}
			gap = thisGap
		}
		return nil
	}
	test := func(original []Datapoint, fixed []Datapoint) error {
		for i := 0; i < len(fixed); i++ {
			if original[i].DateTime.Unix() != fixed[i].DateTime.Unix() {
				return fmt.Errorf("index %d (%d) in original does not match index %d (%d) in fixed\n", i, original[i].DateTime.Unix(), i, fixed[i].DateTime.Unix())
			}
		}
		if len(fixed) != len(original) {
			return fmt.Errorf("did not fill in completely - lens do not match original len: %d, fixed len: %d\n", len(original), len(fixed))
		}
		err := gapCheck(fixed)
		if err != nil {
			return err
		}
		return nil
	}
	{
		original := make([]Datapoint, 0, 0)
		fillInGaps(original, 60)
	}
	{
		original := []Datapoint{
			{Time: "", DateTime: time.Unix(60, 0), Value: 60},
			{Time: "", DateTime: time.Unix(120, 0), Value: 120},
			{Time: "", DateTime: time.Unix(180, 0), Value: 180},
			{Time: "", DateTime: time.Unix(240, 0), Value: 240},
			{Time: "", DateTime: time.Unix(300, 0), Value: 300},
			{Time: "", DateTime: time.Unix(360, 0), Value: 360},
		}
		wrong := []Datapoint{
			{Time: "", DateTime: time.Unix(60, 0), Value: 60},
			{Time: "", DateTime: time.Unix(120, 0), Value: 120},
			// 180 omitted
			// 240 omitted
			// 300 omitted
			{Time: "", DateTime: time.Unix(360, 0), Value: 360},
		}
		fixed := fillInGaps(wrong, 60)
		err := test(original, fixed)
		if err != nil {
			t.Error(err)
		}
	}
	{
		wrong := []Datapoint{
			{Time: "", DateTime: time.Date(2021, 03, 06, 16, 49, 00, 00, time.UTC), Value: 5},
			{Time: "", DateTime: time.Date(2021, 03, 06, 16, 50, 00, 00, time.UTC), Value: 8},
			// 15 minutes omitted
			{Time: "", DateTime: time.Date(2021, 03, 06, 17, 5, 00, 00, time.UTC), Value: 8},
			{Time: "", DateTime: time.Date(2021, 03, 06, 17, 6, 00, 00, time.UTC), Value: 8},
		}
		fixed := fillInGaps(wrong, 60)

		err := gapCheck(fixed)
		if err != nil {
			t.Error(err)
		}
	}
	{
		original := make([]Datapoint, 0, 100)
		for i := 0; i < 60*100; i += 60 {
			original = append(original, Datapoint{
				Time:     "",
				DateTime: time.Unix(int64(i), 0),
				Value:    0,
			})
		}
		messedUp := make([]Datapoint, 100, 100)
		copy(messedUp, original)
		messedUp = append(messedUp[:5], messedUp[5+1:]...)
		messedUp = append(messedUp[:6], messedUp[6+1:]...)
		messedUp = append(messedUp[:7], messedUp[7+1:]...)
		messedUp = append(messedUp[:8], messedUp[8+1:]...)
		messedUp = append(messedUp[:25], messedUp[25+1:]...)
		messedUp = append(messedUp[:67], messedUp[67+1:]...)
		fixed := fillInGaps(messedUp, 60)
		err := test(original, fixed)
		if err != nil {
			t.Error(err)
		}
	}
	{
		var str = `{14:39:00 2021-03-06 14:39:00 +0000 UTC 63} {14:40:00 2021-03-06 14:40:00 +0000 UTC 69} {14:41:00 2021-03-06 14:41:00 +0000 UTC 70} {14:42:00 2021-03-06 14:42:00 +0000 UTC 68} {14:43:00 2021-03-06 14:43:00 +0000 UTC 66} {14:44:00 2021-03-06 14:44:00 +0000 UTC 73} {14:45:00 2021-03-06 14:45:00 +0000 UTC 67} {14:46:00 2021-03-06 14:46:00 +0000 UTC 66} {14:47:00 2021-03-06 14:47:00 +0000 UTC 58} {14:48:00 2021-03-06 14:48:00 +0000 UTC 62} {14:49:00 2021-03-06 14:49:00 +0000 UTC 64} {14:50:00 2021-03-06 14:50:00 +0000 UTC 67} {14:51:00 2021-03-06 14:51:00 +0000 UTC 61} {14:52:00 2021-03-06 14:52:00 +0000 UTC 67} {14:53:00 2021-03-06 14:53:00 +0000 UTC 62} {14:54:00 2021-03-06 14:54:00 +0000 UTC 64} {14:55:00 2021-03-06 14:55:00 +0000 UTC 64} {14:56:00 2021-03-06 14:56:00 +0000 UTC 74} {14:57:00 2021-03-06 14:57:00 +0000 UTC 72} {14:58:00 2021-03-06 14:58:00 +0000 UTC 63} {14:59:00 2021-03-06 14:59:00 +0000 UTC 58} {15:00:00 2021-03-06 15:00:00 +0000 UTC 59} {15:01:00 2021-03-06 15:01:00 +0000 UTC 61} {15:02:00 2021-03-06 15:02:00 +0000 UTC 61} {15:03:00 2021-03-06 15:03:00 +0000 UTC 61} {15:04:00 2021-03-06 15:04:00 +0000 UTC 67} {15:05:00 2021-03-06 15:05:00 +0000 UTC 70} {15:06:00 2021-03-06 15:06:00 +0000 UTC 70} {15:07:00 2021-03-06 15:07:00 +0000 UTC 74} {15:08:00 2021-03-06 15:08:00 +0000 UTC 90} {15:09:00 2021-03-06 15:09:00 +0000 UTC 92} {15:10:00 2021-03-06 15:10:00 +0000 UTC 91} {15:11:00 2021-03-06 15:11:00 +0000 UTC 89} {15:12:00 2021-03-06 15:12:00 +0000 UTC 95} {15:13:00 2021-03-06 15:13:00 +0000 UTC 77} {15:14:00 2021-03-06 15:14:00 +0000 UTC 75} {15:15:00 2021-03-06 15:15:00 +0000 UTC 75} {15:16:00 2021-03-06 15:16:00 +0000 UTC 74} {15:17:00 2021-03-06 15:17:00 +0000 UTC 75} {15:18:00 2021-03-06 15:18:00 +0000 UTC 77} {15:19:00 2021-03-06 15:19:00 +0000 UTC 79} {15:20:00 2021-03-06 15:20:00 +0000 UTC 81} {15:21:00 2021-03-06 15:21:00 +0000 UTC 81} { 2021-03-06 15:22:00 +0000 UTC 81} {15:23:00 2021-03-06 15:23:00 +0000 UTC 77} {15:24:00 2021-03-06 15:24:00 +0000 UTC 77} {15:25:00 2021-03-06 15:25:00 +0000 UTC 77} {15:26:00 2021-03-06 15:26:00 +0000 UTC 77} {15:27:00 2021-03-06 15:27:00 +0000 UTC 77} {15:28:00 2021-03-06 15:28:00 +0000 UTC 81} {15:29:00 2021-03-06 15:29:00 +0000 UTC 81} {15:30:00 2021-03-06 15:30:00 +0000 UTC 81} {15:31:00 2021-03-06 15:31:00 +0000 UTC 83} {15:32:00 2021-03-06 15:32:00 +0000 UTC 86} {15:33:00 2021-03-06 15:33:00 +0000 UTC 90} {15:34:00 2021-03-06 15:34:00 +0000 UTC 92} {15:35:00 2021-03-06 15:35:00 +0000 UTC 87} {15:36:00 2021-03-06 15:36:00 +0000 UTC 87} {15:37:00 2021-03-06 15:37:00 +0000 UTC 89} {15:38:00 2021-03-06 15:38:00 +0000 UTC 81} {15:39:00 2021-03-06 15:39:00 +0000 UTC 76} {15:40:00 2021-03-06 15:40:00 +0000 UTC 75} {15:41:00 2021-03-06 15:41:00 +0000 UTC 77} {15:42:00 2021-03-06 15:42:00 +0000 UTC 76} {15:43:00 2021-03-06 15:43:00 +0000 UTC 79} {15:44:00 2021-03-06 15:44:00 +0000 UTC 81} {15:45:00 2021-03-06 15:45:00 +0000 UTC 81} {15:46:00 2021-03-06 15:46:00 +0000 UTC 81} {15:47:00 2021-03-06 15:47:00 +0000 UTC 94} {15:48:00 2021-03-06 15:48:00 +0000 UTC 119} {15:49:00 2021-03-06 15:49:00 +0000 UTC 131} {15:50:00 2021-03-06 15:50:00 +0000 UTC 135} { 2021-03-06 15:51:00 +0000 UTC 135} {15:52:00 2021-03-06 15:52:00 +0000 UTC 130} { 2021-03-06 15:53:00 +0000 UTC 130} {15:54:00 2021-03-06 15:54:00 +0000 UTC 123} {15:55:00 2021-03-06 15:55:00 +0000 UTC 118} {15:56:00 2021-03-06 15:56:00 +0000 UTC 112} {15:57:00 2021-03-06 15:57:00 +0000 UTC 125} {15:58:00 2021-03-06 15:58:00 +0000 UTC 128} {15:59:00 2021-03-06 15:59:00 +0000 UTC 166} {16:00:00 2021-03-06 16:00:00 +0000 UTC 171} {16:01:00 2021-03-06 16:01:00 +0000 UTC 171} { 2021-03-06 16:02:00 +0000 UTC 171} {16:03:00 2021-03-06 16:03:00 +0000 UTC 175} { 2021-03-06 16:04:00 +0000 UTC 175} {16:05:00 2021-03-06 16:05:00 +0000 UTC 159} {16:06:00 2021-03-06 16:06:00 +0000 UTC 159} {16:07:00 2021-03-06 16:07:00 +0000 UTC 156} { 2021-03-06 16:08:00 +0000 UTC 156} {16:09:00 2021-03-06 16:09:00 +0000 UTC 139} {16:10:00 2021-03-06 16:10:00 +0000 UTC 140} {16:11:00 2021-03-06 16:11:00 +0000 UTC 98} {16:12:00 2021-03-06 16:12:00 +0000 UTC 89} {16:13:00 2021-03-06 16:13:00 +0000 UTC 84} {16:14:00 2021-03-06 16:14:00 +0000 UTC 82} {16:15:00 2021-03-06 16:15:00 +0000 UTC 89} {16:16:00 2021-03-06 16:16:00 +0000 UTC 86} {16:17:00 2021-03-06 16:17:00 +0000 UTC 86} {16:18:00 2021-03-06 16:18:00 +0000 UTC 86} {16:19:00 2021-03-06 16:19:00 +0000 UTC 83} {16:20:00 2021-03-06 16:20:00 +0000 UTC 84} {16:21:00 2021-03-06 16:21:00 +0000 UTC 85} {16:22:00 2021-03-06 16:22:00 +0000 UTC 88} {16:23:00 2021-03-06 16:23:00 +0000 UTC 83} {16:24:00 2021-03-06 16:24:00 +0000 UTC 85} {16:25:00 2021-03-06 16:25:00 +0000 UTC 86} {16:26:00 2021-03-06 16:26:00 +0000 UTC 86} {16:27:00 2021-03-06 16:27:00 +0000 UTC 85} {16:28:00 2021-03-06 16:28:00 +0000 UTC 89} {16:29:00 2021-03-06 16:29:00 +0000 UTC 89} {16:30:00 2021-03-06 16:30:00 +0000 UTC 90} {16:31:00 2021-03-06 16:31:00 +0000 UTC 86} {16:32:00 2021-03-06 16:32:00 +0000 UTC 86} {16:33:00 2021-03-06 16:33:00 +0000 UTC 86} {16:34:00 2021-03-06 16:34:00 +0000 UTC 86} {16:35:00 2021-03-06 16:35:00 +0000 UTC 86} {16:36:00 2021-03-06 16:36:00 +0000 UTC 86} {16:37:00 2021-03-06 16:37:00 +0000 UTC 87} {16:38:00 2021-03-06 16:38:00 +0000 UTC 83} {16:39:00 2021-03-06 16:39:00 +0000 UTC 85} {16:40:00 2021-03-06 16:40:00 +0000 UTC 86} {16:41:00 2021-03-06 16:41:00 +0000 UTC 83} {16:42:00 2021-03-06 16:42:00 +0000 UTC 82} {16:43:00 2021-03-06 16:43:00 +0000 UTC 82} {16:44:00 2021-03-06 16:44:00 +0000 UTC 82} { 2021-03-06 16:45:00 +0000 UTC 82} {16:46:00 2021-03-06 16:46:00 +0000 UTC 88} {16:47:00 2021-03-06 16:47:00 +0000 UTC 85} {16:48:00 2021-03-06 16:48:00 +0000 UTC 87} {16:49:00 2021-03-06 16:49:00 +0000 UTC 85} {16:50:00 2021-03-06 16:50:00 +0000 UTC 85} { 2021-03-06 16:51:00 +0000 UTC 85} {17:06:00 2021-03-06 17:06:00 +0000 UTC 81} {17:07:00 2021-03-06 17:07:00 +0000 UTC 81} {17:08:00 2021-03-06 17:08:00 +0000 UTC 83} {17:09:00 2021-03-06 17:09:00 +0000 UTC 84} {17:10:00 2021-03-06 17:10:00 +0000 UTC 86} {17:11:00 2021-03-06 17:11:00 +0000 UTC 84} {17:12:00 2021-03-06 17:12:00 +0000 UTC 84} {17:13:00 2021-03-06 17:13:00 +0000 UTC 80} {17:14:00 2021-03-06 17:14:00 +0000 UTC 82} {17:15:00 2021-03-06 17:15:00 +0000 UTC 83} {17:16:00 2021-03-06 17:16:00 +0000 UTC 83} {17:17:00 2021-03-06 17:17:00 +0000 UTC 80} {17:18:00 2021-03-06 17:18:00 +0000 UTC 83} {17:19:00 2021-03-06 17:19:00 +0000 UTC 81} {17:20:00 2021-03-06 17:20:00 +0000 UTC 82} {17:21:00 2021-03-06 17:21:00 +0000 UTC 85} {17:22:00 2021-03-06 17:22:00 +0000 UTC 89} {17:23:00 2021-03-06 17:23:00 +0000 UTC 87} {17:24:00 2021-03-06 17:24:00 +0000 UTC 92} {17:25:00 2021-03-06 17:25:00 +0000 UTC 87} {17:26:00 2021-03-06 17:26:00 +0000 UTC 87} {17:27:00 2021-03-06 17:27:00 +0000 UTC 89} {17:28:00 2021-03-06 17:28:00 +0000 UTC 83} {17:29:00 2021-03-06 17:29:00 +0000 UTC 81} {17:30:00 2021-03-06 17:30:00 +0000 UTC 83} {17:31:00 2021-03-06 17:31:00 +0000 UTC 82} {17:32:00 2021-03-06 17:32:00 +0000 UTC 81} {17:33:00 2021-03-06 17:33:00 +0000 UTC 82} {17:34:00 2021-03-06 17:34:00 +0000 UTC 80} {17:35:00 2021-03-06 17:35:00 +0000 UTC 87} {17:36:00 2021-03-06 17:36:00 +0000 UTC 79} {17:37:00 2021-03-06 17:37:00 +0000 UTC 83} {17:38:00 2021-03-06 17:38:00 +0000 UTC 84} {17:39:00 2021-03-06 17:39:00 +0000 UTC 85} {17:40:00 2021-03-06 17:40:00 +0000 UTC 82} {17:41:00 2021-03-06 17:41:00 +0000 UTC 84} {17:42:00 2021-03-06 17:42:00 +0000 UTC 85} {17:43:00 2021-03-06 17:43:00 +0000 UTC 82} {17:44:00 2021-03-06 17:44:00 +0000 UTC 85} {17:45:00 2021-03-06 17:45:00 +0000 UTC 87} {17:46:00 2021-03-06 17:46:00 +0000 UTC 86} {17:47:00 2021-03-06 17:47:00 +0000 UTC 82} {17:48:00 2021-03-06 17:48:00 +0000 UTC 82} {17:49:00 2021-03-06 17:49:00 +0000 UTC 86} {17:50:00 2021-03-06 17:50:00 +0000 UTC 79} {17:51:00 2021-03-06 17:51:00 +0000 UTC 81} {17:52:00 2021-03-06 17:52:00 +0000 UTC 83} {17:53:00 2021-03-06 17:53:00 +0000 UTC 81} {17:54:00 2021-03-06 17:54:00 +0000 UTC 81} {17:55:00 2021-03-06 17:55:00 +0000 UTC 78} {17:56:00 2021-03-06 17:56:00 +0000 UTC 81} {17:57:00 2021-03-06 17:57:00 +0000 UTC 81} {17:58:00 2021-03-06 17:58:00 +0000 UTC 84} {17:59:00 2021-03-06 17:59:00 +0000 UTC 85} {18:00:00 2021-03-06 18:00:00 +0000 UTC 85} {18:01:00 2021-03-06 18:01:00 +0000 UTC 86} {18:02:00 2021-03-06 18:02:00 +0000 UTC 89} {18:03:00 2021-03-06 18:03:00 +0000 UTC 87} {18:04:00 2021-03-06 18:04:00 +0000 UTC 88} {18:05:00 2021-03-06 18:05:00 +0000 UTC 91} {18:06:00 2021-03-06 18:06:00 +0000 UTC 86} {18:07:00 2021-03-06 18:07:00 +0000 UTC 86} {18:08:00 2021-03-06 18:08:00 +0000 UTC 84} {18:09:00 2021-03-06 18:09:00 +0000 UTC 83} {18:10:00 2021-03-06 18:10:00 +0000 UTC 85} {18:11:00 2021-03-06 18:11:00 +0000 UTC 83} {18:12:00 2021-03-06 18:12:00 +0000 UTC 83} {18:13:00 2021-03-06 18:13:00 +0000 UTC 83} {18:14:00 2021-03-06 18:14:00 +0000 UTC 84} {18:15:00 2021-03-06 18:15:00 +0000 UTC 87} {18:16:00 2021-03-06 18:16:00 +0000 UTC 84} {18:17:00 2021-03-06 18:17:00 +0000 UTC 84} {18:18:00 2021-03-06 18:18:00 +0000 UTC 79} {18:19:00 2021-03-06 18:19:00 +0000 UTC 83} {18:20:00 2021-03-06 18:20:00 +0000 UTC 81} {18:21:00 2021-03-06 18:21:00 +0000 UTC 77} {18:22:00 2021-03-06 18:22:00 +0000 UTC 77} {18:23:00 2021-03-06 18:23:00 +0000 UTC 79} {18:24:00 2021-03-06 18:24:00 +0000 UTC 81} {18:25:00 2021-03-06 18:25:00 +0000 UTC 78} {18:26:00 2021-03-06 18:26:00 +0000 UTC 81}`
		s := strings.Split(str, "} {")
		dataset := make([]Datapoint, 0, 220)
		for _, i2 := range s {
			val, tim := parseTime(t, i2)
			dataset = append(dataset, Datapoint{
				Time:     "",
				DateTime: tim,
				Value:    val,
			})
		}
		fixed := fillInGaps(dataset, 60)
		err := gapCheck(fixed)
		if err != nil {
			t.Error(err)
		}
	}
}

func parseTime(t *testing.T, timestamp string) (int, time.Time) {
	a := strings.Trim(timestamp, "}{")
	ss := strings.Split(a, " ")
	tzStr := fmt.Sprintf("%sT%s", ss[1], ss[2])
	val, _ := strconv.Atoi(ss[5])
	tim, err := time.Parse("2006-01-02T15:04:05", tzStr)
	if err != nil {
		t.Error(err)
	}
	return val, tim
}
