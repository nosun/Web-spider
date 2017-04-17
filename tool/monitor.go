package tool

import (
	"errors"
	"time"

	"fmt"

	"runtime"

	sched "github.com/yanchenxu/Web-spider/scheduler"
)

//Record 日志记录函数
type Record func(level byte, contet string)

//Monitoring 调度器监控函数
func Monitoring(
	scheduler sched.Scheduler, //监控目标
	intervalNs time.Duration, //检查间隔时间，纳秒
	maxIdleCount uint, //代表最大空闲计数
	autoStop bool, //是否在调度器空闲一段时间后，自行停止调度器
	detailSummary bool, //用来表示是否需要详细的摘要信息
	record Record, //日志记录函数
) <-chan uint64 {
	if scheduler == nil {
		panic(errors.New("The scheduler is invalid"))
	}
	//防止过小的参数数值对爬取流程的影响
	if intervalNs < time.Millisecond {
		intervalNs = time.Millisecond
	}

	if maxIdleCount < 1000 {
		maxIdleCount = 1000
	}

	//监控停止通知器
	stopNotifier := make(chan byte, 1)
	//接受和报告错误
	reportError(scheduler, record, stopNotifier)
	//记录摘要信息
	recordSummary(scheduler, detailSummary, record, stopNotifier)
	//检查计数通道
	checkCountChan := make(chan uint64, 2)
	//检查空间状态
	checkStatus(scheduler,
		intervalNs,
		maxIdleCount,
		autoStop,
		checkCountChan, record,
		stopNotifier)
	return checkCountChan
}

func reportError(scheduler sched.Scheduler,
	record Record,
	stopNotifier <-chan byte) {
	go func() {
		//等待调度器开启
		waitForSchedulerStart(scheduler)
		for {
			//查看监控停止通知器
			select {
			case <-stopNotifier:
				return
			default:
			}
			errorChan := scheduler.ErrorChan()
			if errorChan == nil {
				return
			}
			err := <-errorChan
			if err != nil {
				errMsg := fmt.Sprintf("Error (received from error channel): %s", err)
				record(2, errMsg)
			}
			time.Sleep(time.Microsecond)
		}
	}()
}

// 等待调度器开启
func waitForSchedulerStart(scheduler sched.Scheduler) {
	for !scheduler.Running() {
		time.Sleep(time.Microsecond)
	}
}

//摘要信息模板
var summaryForMonitoring = "Monitor - Collected information[%d]:\n" +
	"	Goroutine number: %d\n" +
	"	Scheduler:\n%s" +
	"	Escaped time: %s\n"

//记录摘要信息
func recordSummary(
	scheduler sched.Scheduler,
	detailSummary bool,
	record Record,
	stopNotifier <-chan byte) {
	var recordCount uint64 = 1
	startTime := time.Now()
	var prevSchedSummary sched.SchedSummary
	var prevNumGoroutine int

	for {
		//查看监控停止通知器
		select {
		case <-stopNotifier:
			return
		default:
			//获取摘要信息的各组成部分
			currNumGOroutine := runtime.NumGoroutine()
			currSchedSummary := scheduler.Summary(" ")
			//比对前后两份摘要信息的一致性。只有不一致才会予以记录
			if currNumGOroutine != prevNumGoroutine ||
				!currSchedSummary.Same(prevSchedSummary) {
				schedSummaryStr := func() string {
					if detailSummary {
						return currSchedSummary.Detail()
					}
					return currSchedSummary.String()

				}()
				info := fmt.Sprintf(summaryForMonitoring,
					recordCount,
					currNumGOroutine,
					schedSummaryStr,
					time.Since(startTime).String(),
				)
				record(0, info)
				prevNumGoroutine = currNumGOroutine
				prevSchedSummary = currSchedSummary
				recordCount++
				time.Sleep(time.Microsecond)
			}
		}
	}
}

//已达到最大空间技术的消息模板
var msgReachMaxIdleCount = "The schedler has been idle for a period of time " +
	"(about %s)." +
	"Now consider what stop it."

//检查状态，并在满足持续空闲时间的条件时采取必要的措施
func checkStatus(scheduler sched.Scheduler,
	intervalNs time.Duration,
	maxIdleCount uint,
	autoStop bool,
	checkCountChan chan<- uint64,
	record Record,
	stopNotifier chan<- byte) {
	var checkCount uint64
	go func() {
		defer func() {
			stopNotifier <- 1
			stopNotifier <- 2
			checkCountChan <- checkCount
		}()
		var idleCount uint
		var firstIdleTime time.Time
		for {
			//检查调度器的空闲状态
			if scheduler.Idle() {
				idleCount++
				if idleCount == 1 {
					firstIdleTime = time.Now()
				}

				if idleCount >= maxIdleCount {
					msg := fmt.Sprintf(msgReachMaxIdleCount, time.Since(firstIdleTime).String())
					record(0, msg)

					//再次检查调度器的空闲状态，确保它已经可以被停止
					if scheduler.Idle() {
						if autoStop {
							var result string
							if scheduler.Stop() {
								result = "success"
							} else {
								result = "failing"
							}
							msg = fmt.Sprintf("stop scheduler %s\n", result)
							record(0, msg)
						}
						break
					} else {
						if idleCount > 0 {
							idleCount = 0
						}
					}
				}
			} else {
				if idleCount > 0 {
					idleCount = 0
				}
			}
			checkCount++
			time.Sleep(intervalNs)
		}
	}()
}
