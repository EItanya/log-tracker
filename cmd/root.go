package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	filepathArg   string
	numberOfLines int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "log-tracker",
	Short: "Pretty version of unix tail command",

	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE:              rootRunE,
	PersistentPreRunE: rootPreRunE,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".cobra.yaml", "config file (default is .cobra.yaml)")
	rootCmd.PersistentFlags().StringVar(&filepathArg, "filepath", "", "filepath of file to be tailed")
	rootCmd.PersistentFlags().IntVarP(&numberOfLines, "number", "n", 10, "number of lines to print from console")
	rootCmd.PersistentFlags().BoolP("follow", "f", false, "Output to stdout as new lines are written")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("follow", "f", false, "Output to stdout as new lines are written")
	// viper.BindPFlag("follow", rootCmd.PersistentFlags().Lookup("follow"))
	viper.BindPFlags(rootCmd.PersistentFlags())

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		viper.AddConfigPath(".")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func rootPreRunE(cmd *cobra.Command, args []string) error {
	if _, err := exec.Command("tail").Output(); err != nil {
		if cmdErr, ok := err.(*exec.Error); ok {
			log.Println("valid tail command found on current system")
			return cmdErr
		}
	}
	if err := checkFilePath(args); err != nil {
		return err
	}
	return nil
}

func checkFilePath(args []string) error {
	if len(args) == 1 {
		if _, err := os.Stat(args[0]); err == nil {
			filepathArg = args[0]
			return nil
		}
		return fmt.Errorf("('%s') is not a valid filepath", args[0])
	}
	if filepathArg != "" {
		if _, err := os.Stat(filepathArg); err == nil {
			return nil
		}
		return fmt.Errorf("('%s') is not a valid filepath", filepathArg)
	}
	return errors.New("Filepath not provided, can either be provided via flag or the first arg")
}

func rootRunE(cmd *cobra.Command, args []string) error {
	follow := viper.GetBool("follow")

	tailCmdStringArgs := buildTailCommandArgs(follow, numberOfLines, filepathArg)
	tailCmd := exec.Command("tail", tailCmdStringArgs...)
	if follow {
		log.Println("Beginning follow mode")
		err := followMode(tailCmd)
		if err != nil {
			return err
		}
	} else {
		log.Println("Logging default amount")
		err := standardMode(tailCmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildTailCommandArgs(follow bool, number int, filepath string) []string {
	result := make([]string, 0)
	if follow {
		result = append(result, "-f")
	}
	result = append(result, []string{"-n", fmt.Sprintf("%d", number), filepath}...)
	return result
}

func followMode(tailCmd *exec.Cmd) error {
	_, err := tailCmd.StderrPipe()
	if err != nil {
		return err
	}
	_, err = tailCmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = tailCmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func standardMode(tailCmd *exec.Cmd) error {
	out, err := tailCmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprint(err) + " : " + string(out))
	}
	fmt.Println(string(out))
	return nil
}
