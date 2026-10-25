package local

import (
	"time"

	"github.com/GabeCordo/DistributedFunctions/internal/shared/terminal"
)

type BannerFactory struct {
	defaultColour string
}

func NewBannerFactory() BannerFactory {
	return BannerFactory{
		defaultColour: terminal.Purple,
	}
}

func (bf BannerFactory) Create(t time.Time) string {

	optionalBanner := ""
	if (t.Month() == time.October) && (t.Day() > 20) {
		bf.defaultColour = terminal.Orange
		optionalBanner = bf.halloweenBanner()
	} else if (t.Month() == time.December) && (t.Day() <= 25) {
		bf.defaultColour = terminal.Green
		optionalBanner = bf.christmasBanner()
	} else if ((t.Month() == time.December) && (t.Day() > 25)) || ((t.Month() == time.January) && (t.Day() <= 1)) {
		bf.defaultColour = terminal.Yellow
		optionalBanner = bf.newYearsBanner()
	}

	return bf.standardHeader() + "\n" + optionalBanner + "\n" + bf.standardFooter()
}

func (bf BannerFactory) standardHeader() string {
	return bf.defaultColour + "    ___" + terminal.Reset + " _     _        _ _           _           _   " + bf.defaultColour + "___                 " + terminal.Reset + "_   _  " + "\n" +
		bf.defaultColour + "   /   " + terminal.Reset + "(_)___| |_ _ __(_) |__  _   _| |_ ___  __| | " + bf.defaultColour + "/ __\\   _ _ __   ___" + terminal.Reset + "| |_(_) ___  _ __  ___  " + "\n" +
		bf.defaultColour + "  / /\\ /" + terminal.Reset + " / __| __| '__| | '_ \\| | | | __/ _ \\/ _` |" + bf.defaultColour + "/ _\\| | | | '_ \\ / __" + terminal.Reset + "| __| |/ _ \\| '_ \\/ __|" + "\n" +
		bf.defaultColour + " / /_//" + terminal.Reset + "| \\__ \\ |_| |  | | |_) | |_| | ||  __/ (_| " + bf.defaultColour + "/ /  | |_| | | | | (__" + terminal.Reset + "| |_| | (_) | | | \\__ \\" + "\n" +
		bf.defaultColour + "/___,'" + terminal.Reset + " |_|___/\\__|_|  |_|_.__/ \\__,_|\\__\\___|\\__,_" + bf.defaultColour + "\\/    \\__,_|_| |_|\\___" + terminal.Reset + "|\\__|_|\\___/|_| |_|___/" + "\n"
}

func (bf BannerFactory) standardFooter() string {
	return "[+] " + bf.defaultColour + "Distributed Functions " + terminal.Reset + "\n" +
		"[+]" + bf.defaultColour + " by Gabriel Cordovado Copyright 2022-2026" + terminal.Reset + "\n"
}

func (bf BannerFactory) halloweenBanner() string {
	return "      \\\\                \\\\                \\\\\n" +
		" .-'```^```'-.     .-'```^```'-.     .-'```^```'-.\n" +
		"/   /\\ __ /\\  \\   /   (\\ __ /)  \\   /   /) __ (\\  \\\n" +
		"|   " + bf.defaultColour + "^^" + terminal.Reset + " \\/ " + bf.defaultColour + "^^" + terminal.Reset + "  |   |    " + bf.defaultColour + "`" + terminal.Reset + " \\/ " + bf.defaultColour + "`" + terminal.Reset + "   |   |   " + bf.defaultColour + "^" + terminal.Reset + "  \\/  " + bf.defaultColour + "^" + terminal.Reset + "  |\n" +
		"\\   \\_.__._/  /   \\    \\____/   /   \\    `'=='`   /\n" +
		" `'-.......-'`     `'-.......-'`     `'-.......-'` ldb\n"
}

func (bf BannerFactory) christmasBanner() string {
	return "          *\n" +
		"         /" + terminal.Red + "." + terminal.Reset + "\\\n" +
		"        /" + terminal.Blue + "." + terminal.Yellow + "." + terminal.Reset + "\\\n" +
		"        /'" + terminal.Red + "." + terminal.Reset + "'\\\n" +
		"       /" + terminal.Yellow + "." + terminal.Reset + "''" + terminal.Red + "." + terminal.Reset + "'\\\n" +
		"       /" + terminal.Orange + "." + terminal.Reset + "'" + terminal.Blue + "." + terminal.Reset + "'" + terminal.Yellow + "." + terminal.Reset + "\\\n" +
		"\"'\"\"\"\"/'" + terminal.Red + "." + terminal.Reset + "''" + terminal.Orange + "." + terminal.Reset + "'" + terminal.Blue + "." + terminal.Reset + "\\\"\"'\"'\"\n" +
		"  jgs ^^^[_]^^^\n"
}

func (bf BannerFactory) newYearsBanner() string {
	return "                                   .''.\n" +
		"       .''.      .        *''*    :_\\/_:     .\n" +
		"      :_\\/_:   _\\(/_  .:.*_\\/_*   : /\\ :  .'.:.'.\n" +
		":_\\/_:'.:::.  | ' *''*    * '.\\'/.'_\\(/_ '.':'.'\n" +
		" : /\\ : :::::  =  *_\\/_*     -= o =- /)\\     '  *\n" +
		"  '..'  ':::' === * /\\ *     .'/.\\'.  ' ._____\n" +
		"      *        |   *..*         :       |.   |' .---\"|\n" +
		"        *      |     _           .--'|  ||   | _|    |\n" +
		"        *      |  .-'|       __  |   |  |    ||      |\n" +
		"     .-----.   |  |' |  ||  |  | |   |  |    ||      |\n" +
		" ___'       ' /\"\\ |  '-.\"\".    '-'   '-.'    '`      |____\n" +
		"jgs~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n" +
		"  &                    ~-~-~-~-~-~-~-~-~-~   /|\n" +
		" ejm97    )      ~-~-~-~-~-~-~-~  /|~       /_|\\\n" +
		"        _-H-__  -~-~-~-~-~-~     /_|\\    -~======-~\n" +
		"~-\\XXXXXXXXXX/~     ~-~-~-~     /__|_\\ ~-~-~-~\n" +
		"~-~-~-~-~-~    ~-~~-~-~-~-~    ========  ~-~-~-~\n" +
		"      ~-~~-~-~-~-~-~-~-~-~-~-~-~ ~-~~-~-~-~-~\n" +
		"                        ~-~~-~-~-~-~\n"
}
