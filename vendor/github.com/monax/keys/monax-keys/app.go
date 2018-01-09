package keys

import (
	"fmt"
	"os"

	"github.com/monax/keys/common"

	"github.com/spf13/cobra"
)

func init() {

	// note these are only for use by the client
	if keysHost := os.Getenv("MONAX_KEYS_HOST"); keysHost != "" {
		DefaultHost = keysHost
	}
	if keysPort := os.Getenv("MONAX_KEYS_PORT"); keysPort != "" {
		DefaultPort = keysPort
	}
}

var (
	DefaultKeyType  = "ed25519,ripemd160"
	DefaultDir      = common.KeysPath
	DefaultHashType = "sha256"

	DefaultHost = "localhost"
	DefaultPort = "4767"
	TestPort    = "7674"
	TestAddr    = "http://" + DefaultHost + ":" + TestPort

	// set in before()
	DaemonAddr string

	/* flag vars */
	//global
	logLevel int // currently only info level available; ignored
	KeysDir  string
	KeyName  string
	KeyAddr  string
	KeyHost  string
	KeyPort  string

	//keygenCmd only
	NoPassword bool
	KeyType    string

	//hashCmd only
	HashType string
	HexByte  bool

	// lockCmd only
	UnlockTime int // minutes

	Verbose bool
	Debug   bool
)

var EKeys = &cobra.Command{
	Use:   "monax-keys",
	Short: "Generate and manage keys for producing signatures",
	Long:  "A tool for doing a bunch of cool stuff with keys.",
	Run:   func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func Execute() {
	BuildKeysCommand()
	EKeys.PersistentPreRun = before
	EKeys.Execute()
}

func BuildKeysCommand() {
	nameCmd.AddCommand(nameRmCmd, nameLsCmd)

	EKeys.AddCommand(keygenCmd)
	EKeys.AddCommand(lockCmd)
	EKeys.AddCommand(unlockCmd)
	EKeys.AddCommand(nameCmd)
	EKeys.AddCommand(signCmd)
	EKeys.AddCommand(pubKeyCmd)
	EKeys.AddCommand(verifyCmd)
	EKeys.AddCommand(hashCmd)
	EKeys.AddCommand(serverCmd)
	EKeys.AddCommand(importCmd)
	EKeys.AddCommand(convertCmd)
	addKeysFlags()
}

var keygenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate a key",
	Long:  "Generates a key using (insert crypto pkgs used)",
	Run:   cliKeygen,
}

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "lock a key",
	Long:  "lock an unlocked key by re-encrypting it",
	Run:   cliLock,
}

var unlockCmd = &cobra.Command{
	Use:   "unlock",
	Short: "unlock a key",
	Long:  "unlock an unlocked key by supplying the password",
	Run:   cliUnlock,
}

var nameCmd = &cobra.Command{
	Use:   "name",
	Short: "Manage key names. `monax-keys name <name> <address>`",
	Long:  "Manage key names. `monax-keys name <name> <address>`",
	Run:   cliName,
}

var nameLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "list key names",
	Long:  "list key names",
	Run:   cliNameLs,
}

var nameRmCmd = &cobra.Command{
	Use:   "rm",
	Short: "rm key name",
	Long:  "rm key name",
	Run:   cliNameRm,
}

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "monax-keys sign --addr <address> <hash>",
	Long:  "monax-keys sign --addr <address> <hash>",
	Run:   cliSign,
}

var pubKeyCmd = &cobra.Command{
	Use:   "pub",
	Short: "monax-keys pub --addr <addr>",
	Long:  "monax-keys pub --addr <addr>",
	Run:   cliPub,
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "monax-keys verify --addr <addr> <hash> <sig>",
	Long:  "monax-keys verify --addr <addr> <hash> <sig>",
	Run:   cliVerify,
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "monax-keys convert --addr <address>",
	Long:  "monax-keys convert --addr <address>",
	Run:   cliConvert,
}

var hashCmd = &cobra.Command{
	Use:   "hash",
	Short: "monax-keys hash <some data>",
	Long:  "monax-keys hash <some data>",
	Run:   cliHash,
}
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "monax-keys server",
	Long:  "monax-keys server",
	Run:   cliServer,
}
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "monax-keys import <priv key> | /path/to/keyfile | <key json>",
	Long:  "monax-keys import <priv key> | /path/to/keyfile | <key json>",
	Run:   cliImport,
}

func addKeysFlags() {
	EKeys.PersistentFlags().IntVarP(&logLevel, "log", "l", 0, "specify the location of the directory containing key files")
	EKeys.PersistentFlags().StringVarP(&KeysDir, "dir", "", DefaultDir, "specify the location of the directory containing key files")
	EKeys.PersistentFlags().StringVarP(&KeyName, "name", "", "", "name of key to use")
	EKeys.PersistentFlags().StringVarP(&KeyAddr, "addr", "", "", "address of key to use")
	EKeys.PersistentFlags().StringVarP(&KeyHost, "host", "", DefaultHost, "set the host for talking to the key daemon")
	EKeys.PersistentFlags().StringVarP(&KeyPort, "port", "", DefaultPort, "set the port for key daemon to listen on")
	EKeys.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
	EKeys.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose mode")
	keygenCmd.Flags().StringVarP(&KeyType, "type", "t", DefaultKeyType, "specify the type of key to create. Supports 'secp256k1,sha3' (ethereum),  'secp256k1,ripemd160sha2' (bitcoin), 'ed25519,ripemd160' (tendermint)")
	keygenCmd.Flags().BoolVarP(&NoPassword, "no-pass", "", false, "don't use a password for this key")

	hashCmd.PersistentFlags().StringVarP(&HashType, "type", "t", DefaultHashType, "specify the hash function to use")
	hashCmd.PersistentFlags().BoolVarP(&HexByte, "hex", "", false, "the input should be hex decoded to bytes first")

	importCmd.PersistentFlags().StringVarP(&KeyType, "type", "t", DefaultKeyType, "import a key")
	importCmd.Flags().BoolVarP(&NoPassword, "no-pass", "", false, "don't use a password for this key")

	verifyCmd.PersistentFlags().StringVarP(&KeyType, "type", "t", DefaultKeyType, "key type")

	unlockCmd.PersistentFlags().IntVarP(&UnlockTime, "time", "t", 10, "number of minutes to unlock key for. defaults to 10, 0 for forever")
}

func checkMakeDataDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func before(cmd *cobra.Command, args []string) {

	DaemonAddr = fmt.Sprintf("http://%s:%s", KeyHost, KeyPort)
}

func LogToChannel(answer []byte) {
	fmt.Fprintln(os.Stdout, string(answer))
}
