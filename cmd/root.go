package cmd

import (
	"fmt"
	"os"

	"github.com/rizkianakbar/kbrprime-be/config"
	"github.com/rizkianakbar/kbrprime-be/internal/app/appcontext"
	"github.com/rizkianakbar/kbrprime-be/internal/app/commons"
	"github.com/rizkianakbar/kbrprime-be/internal/app/repository"
	"github.com/rizkianakbar/kbrprime-be/internal/app/repository/healtyRepository"
	"github.com/rizkianakbar/kbrprime-be/internal/app/server"
	"github.com/rizkianakbar/kbrprime-be/internal/app/service"
	"github.com/rizkianakbar/kbrprime-be/internal/app/service/healtyService"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	gologger "github.com/mo-taufiq/go-logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kbrprime-api",
	Short: "A brief description of your application",
	Long:  `A longer description that spans multiple lines and likely contains examples and usage of using your application.`,
	Run: func(cmd *cobra.Command, args []string) {
		loadEnv("")
		start()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}

func initCommonOptions() (options commons.Options, err error) {
	cfg := config.Config()
	app := appcontext.NewAppContext(cfg)

	logLevel := zerolog.InfoLevel
	logLevelP, err := zerolog.ParseLevel(os.Getenv("APP_LOG_LEVEL"))
	if err == nil {
		logLevel = logLevelP
	}
	zerolog.SetGlobalLevel(logLevel)

	validator := validator.New()

	var mysqlDB *gorm.DB
	if app.GetMysqlOption().IsEnable {
		mysqlDB, err = app.GetDBInstance(appcontext.DBDialectMysql)
		if err != nil {
			log.Info().Msgf("Failed to start, error connect to DB MySQL | %v", err)
			return
		}
	}

	options = commons.Options{
		AppCtx:    app,
		Db:        mysqlDB,
		UUID:      commons.NewUuid(),
		Validator: validator,
	}

	return
}

func loadEnv(envName string) {
	gologger.LogConf.NestedLocationLevel = 2
	log.Logger = log.Output(
		zerolog.ConsoleWriter{
			Out:     os.Stderr,
			NoColor: false,
		},
	)

	dotenvPath := "params/.env"

	if envName == "test" {
		dotenvPath = "params/.env.test"
	}

	err := godotenv.Load(dotenvPath)
	if err != nil {
		log.Error().Msg("Error loading .env file")
	}
}

func start() {
	opt, err := initCommonOptions()
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}

	repo := wiringRepository(repository.Option{
		Options: opt,
	})

	service := wiringService(service.Option{
		Options:      opt,
		Repositories: repo,
	})

	server := server.NewServer(opt, service)

	// run app
	server.StartApp()
}
func wiringRepository(repoOption repository.Option) *repository.Repositories {
	repo := repository.Repositories{
		HealtyRepository: healtyRepository.NewHealtyRepository(repoOption.Db),
	}

	return &repo
}

func wiringService(serviceOption service.Option) *service.Services {
	// wiring up all services
	svc := service.Services{
		HealtyService: healtyService.NewHealtyService(serviceOption.HealtyRepository),
	}
	return &svc
}
