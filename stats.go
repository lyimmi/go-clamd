package clamd

import (
	"bufio"
	"strconv"
	"strings"
)

type Stats struct {
	Pools    int
	State    string
	Threads  Thread
	Memstats MemStat
	Queue    Queue
}

type MemStat struct {
	Heap       float32
	Mmap       float32
	Used       float32
	Free       float32
	Releasable float32
	Pools      int
	PoolsUsed  float32
	PoolsTotal float32
}

type Thread struct {
	Live        int
	Idle        int
	Max         int
	IdleTimeout int
}

type Queue struct {
	Items int
	Stats float32
}

func parseStats(res string) (*Stats, error) {
	stats := Stats{}
	scanner := bufio.NewScanner(strings.NewReader(res))

	var (
		text string
		err  error
	)
	queue := false
	for scanner.Scan() {
		text = scanner.Text()
		if strings.HasPrefix(text, "POOLS") {
			stats.Pools, err = strconv.Atoi(strings.TrimPrefix(text, "POOLS: "))
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(text, "STATE") {
			stats.State = strings.TrimPrefix(text, "STATE: ")
		} else if strings.HasPrefix(text, "THREADS") {
			text = strings.TrimPrefix(text, "THREADS: ")
			split := strings.Split(text, " ")
			prev := ""
			for _, s := range split {
				switch prev {
				case "live":
					stats.Threads.Live, err = strconv.Atoi(s)
					if err != nil {
						return nil, err
					}
				case "idle":
					stats.Threads.Idle, err = strconv.Atoi(s)
					if err != nil {
						return nil, err
					}
				case "max":
					stats.Threads.Max, err = strconv.Atoi(s)
					if err != nil {
						return nil, err
					}
				case "idle-timeout":
					stats.Threads.IdleTimeout, err = strconv.Atoi(s)
					if err != nil {
						return nil, err
					}
				}
				prev = s
			}
		} else if strings.HasPrefix(text, "QUEUE") {
			queue = true
			text = strings.TrimPrefix(strings.Trim(text, "\n\t"), "QUEUE: ")
			stats.Queue.Items, err = strconv.Atoi(strings.Trim(text[:1], " "))
			if err != nil {
				return nil, err
			}
		} else if strings.HasPrefix(text, "\t") && queue {
			text = strings.TrimPrefix(strings.Trim(text, "\t"), "STATS ")
			f, err := strconv.ParseFloat(strings.Trim(text, " "), 32)
			if err != nil {
				return nil, err
			}
			stats.Queue.Stats = float32(f)
		} else if strings.HasPrefix(text, "MEMSTATS:") && queue {
			text = strings.TrimPrefix(text, "MEMSTATS: ")
			split := strings.Split(text, " ")
			prev := ""
			for _, s := range split {
				switch prev {
				case "heap":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.Heap = float32(f)
				case "mmap":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.Mmap = float32(f)
				case "used":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.Used = float32(f)
				case "free":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.Free = float32(f)
				case "releasable":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.Releasable = float32(f)
				case "pools":
					f, err := strconv.Atoi(strings.Trim(s, "M"))
					if err != nil {
						return nil, err
					}
					stats.Memstats.Pools = f
				case "pools_used":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.PoolsUsed = float32(f)
				case "pools_total":
					f, err := strconv.ParseFloat(strings.Trim(s, "M"), 32)
					if err != nil {
						return nil, err
					}
					stats.Memstats.PoolsTotal = float32(f)
				}
				prev = s
			}
		}
	}
	return &stats, nil
}
