package main

import (
	"github.com/ac-zht/gotools/option"
	"github.com/ac-zht/gotools/pool"
	"github.com/ac-zht/super-job/scheduler/internal/repository"
	"github.com/ac-zht/super-job/scheduler/internal/repository/dao"
	"github.com/ac-zht/super-job/scheduler/internal/service"
	"github.com/ac-zht/super-job/scheduler/internal/service/http/client"
	"github.com/ac-zht/super-job/scheduler/internal/service/notify"
	"github.com/ac-zht/super-job/scheduler/internal/web"
	"github.com/ac-zht/super-job/scheduler/ioc"
	"github.com/google/wire"
	"net/http"
)

func NewHttpClient() *client.HttpClient {
	return &client.HttpClient{
		Url:     "",
		Timeout: 5,
		Client:  &http.Client{},
		Req:     &http.Request{},
	}
}

func NewNotifiableSlice(repo repository.SettingRepository) []notify.Notifiable {
	return []notify.Notifiable{
		notify.NewMailNotify(repo),
		notify.NewSlackNotify(repo, NewHttpClient()),
		notify.NewWebhookNotify(repo),
	}
}

var notifyServiceProvider = wire.NewSet(
	dao.NewSettingDAO,
	repository.NewSettingRepository,
	NewNotifiableSlice,
	NewNotifyService,
)

func NewNotifyService(nts ...notify.Notifiable) notify.Service {
	return notify.NewService(100, nts...)
}

func NewOnDemandBlockTaskPool() *pool.OnDemandBlockTaskPool {
	quickPool, _ := pool.NewOnDemandBlockTaskPool(5, 10)
	return quickPool
}

func NewCronJobServiceOption() []option.Option[service.CronJobService] {
	return []option.Option[service.CronJobService]{func(s *service.CronJobService) {}}
}

func NewSchedulerOption() []option.Option[web.Scheduler] {
	return []option.Option[web.Scheduler]{func(s *web.Scheduler) {}}
}

func InitScheduler() *web.Scheduler {
	wire.Build(
		ioc.InitDB,

		dao.NewTaskDAO,
		dao.NewTaskLogDAO,

		repository.NewTaskRepository,
		repository.NewTaskLogRepository,

		notifyServiceProvider,

		NewCronJobServiceOption,
		service.NewJobService,

		NewOnDemandBlockTaskPool,
		NewSchedulerOption,
		web.NewScheduler,
	)
	return &web.Scheduler{}
}
