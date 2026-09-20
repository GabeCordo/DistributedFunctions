package controllers

import (
	"fmt"
	"log"
	"os"

	commandline "github.com/GabeCordo/Commandline"
	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
	"github.com/GabeCordo/DistributedFunctions/internal/targets/core"
)

type StartCommand struct {
}

func (sc StartCommand) banner() {
	banner := "    ___ _     _        _ _           _           _   ___                 _   _  \n" +
		"   /   (_)___| |_ _ __(_) |__  _   _| |_ ___  __| | / __\\   _ _ __   ___| |_(_) ___  _ __  ___  \n" +
		"  / /\\ / / __| __| '__| | '_ \\| | | | __/ _ \\/ _` |/ _\\| | | | '_ \\ / __| __| |/ _ \\| '_ \\/ __|\n" +
		" / /_//| \\__ \\ |_| |  | | |_) | |_| | ||  __/ (_| / /  | |_| | | | | (__| |_| | (_) | | | \\__ \\\n" +
		"/___,' |_|___/\\__|_|  |_|_.__/ \\__,_|\\__\\___|\\__,_\\/    \\__,_|_| |_|\\___|\\__|_|\\___/|_| |_|___/"
	fmt.Println(banner)
	fmt.Println("[+] " + terminal.Purple + "Distributed Functions " + terminal.Reset)
	fmt.Println("[+]" + terminal.Purple + " by Gabriel Cordovado 2022-2026" + terminal.Reset)
	fmt.Println()
}

const MongoDatabaseUriEnv = "MONGO_DATABASE_URI"

type EnvironmentVariables struct {
	MongoDbUri string
}

func (sc StartCommand) readEnvironmentVariables() (env EnvironmentVariables) {

	env.MongoDbUri = os.Getenv(MongoDatabaseUriEnv)
	return env
}

func (sc StartCommand) verifyMandatoryEnvironmentVariables(env EnvironmentVariables) (err error) {

	// there are no mandatory environment variables for now.
	err = nil
	return err
}

func (sc StartCommand) Run(cli *commandline.CommandLine) commandline.TerminateOnCompletion {

	// check to see that the etl thread has been initialized with the required files
	// if it has not, fail and tell the operator to call the 'etl init' command
	if _, err := os.Stat(DefaultConfigsFolder); err != nil {
		fmt.Printf("missing configurations folder at %s\nmake sure you run 'DistributedFunctions init'\n", DefaultConfigsFolder)
		return commandline.Terminate
	}

	sc.banner()

	env := sc.readEnvironmentVariables()
	err := sc.verifyMandatoryEnvironmentVariables(env)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	cfg, err := core.GetConfigInstance(DefaultCoreConfigFile)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}
	// Override the database URL if the environment variable is set
	if env.MongoDbUri != "" {
		cfg.Database.Url = env.MongoDbUri
	}

	c, err := core.New(cfg)
	if err != nil {
		log.Println(err)
		return commandline.Terminate
	}

	c.Run()

	return commandline.Terminate
}
