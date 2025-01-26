package files

const CODERUNNER_DIR = ".coderunner"

var defaultIgnorePatterns = []string{
	".git",
	".DS_Store",
	"._.DS_Store",
	"Thumbs.db",
	"desktop.ini",
	"*.swp",        // vim swap files
	"*~",           // temp files
	".vscode",      // IDE files
	".idea",        // IDE files
	"*.tmp",        // temp files
	"*.temp",       // temp files
	".env",         // environment files
	"node_modules", // node modules
	".coderunner",  // Coderunner files
}
