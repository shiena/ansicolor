package ansicolor

// TextAttribute is a wrapper around ANSI codes
type TextAttribute string

// AllAttributesOff turns all attributes off (resets the colors to the startup's defaults)
const AllAttributesOff TextAttribute = "\x1b[0m"

// Basic text attributes

// BoldOn enables bold text
const BoldOn TextAttribute = "\x1b[1m"

// BoldOff disables bold text
const BoldOff TextAttribute = "\x1b[21m"

// BlinkOn enables blinking text
const BlinkOn TextAttribute = "\x1b[5m"

// BlinkOff disables blinking text
const BlinkOff TextAttribute = "\x1b[25m"

// UnderlineOn enables text underlining
const UnderlineOn TextAttribute = "\x1b[4m"

// UnderlineOff disables text underlining
const UnderlineOff TextAttribute = "\x1b[24m"

// Foreground attributes

// FgBlack sets the foreground color (text color) to Black
const FgBlack TextAttribute = "\x1b[30m"

// FgRed sets the foreground color (text color) to Red
const FgRed TextAttribute = "\x1b[31m"

// FgGreen sets the foreground color (text color) to Green
const FgGreen TextAttribute = "\x1b[32m"

// FgYellow sets the foreground color (text color) to Yellow
const FgYellow TextAttribute = "\x1b[33m"

// FgBlue sets the foreground color (text color) to Blue
const FgBlue TextAttribute = "\x1b[34m"

// FgMagenta sets the foreground color (text color) to Magenta
const FgMagenta TextAttribute = "\x1b[35m"

// FgCyan sets the foreground color (text color) to Cyan
const FgCyan TextAttribute = "\x1b[36m"

// FgWhite sets the foreground color (text color) to White
const FgWhite TextAttribute = "\x1b[37m"

// FgDefault resets the foreground color (text color) to its defaults at startup
const FgDefault TextAttribute = "\x1b[39m"

// FgLightGray sets the foreground color (text color) to Light Gray
const FgLightGray TextAttribute = "\x1b[90m"

// FgLightRed sets the foreground color (text color) to Light Red
const FgLightRed TextAttribute = "\x1b[91m"

// FgLightGreen sets the foreground color (text color) to Light Green
const FgLightGreen TextAttribute = "\x1b[92m"

// FgLightYellow sets the foreground color (text color) to Light Yellow
const FgLightYellow TextAttribute = "\x1b[93m"

// FgLightBlue sets the foreground color (text color) to Light Blue
const FgLightBlue TextAttribute = "\x1b[94m"

// FgLightMagenta sets the foreground color (text color) to Light Magenta
const FgLightMagenta TextAttribute = "\x1b[95m"

// FgLightCyan sets the foreground color (text color) to Light Cyan
const FgLightCyan TextAttribute = "\x1b[96m"

// FgLightWhite sets the foreground color (text color) to Light White
const FgLightWhite TextAttribute = "\x1b[97m"

// Background attributes

// BgBlack sets the foreground color (text color) to Black
const BgBlack TextAttribute = "\x1b[40m"

// BgRed sets the foreground color (text color) to Red
const BgRed TextAttribute = "\x1b[41m"

// BgGreen sets the foreground color (text color) to Green
const BgGreen TextAttribute = "\x1b[42m"

// BgYellow sets the foreground color (text color) to Yellow
const BgYellow TextAttribute = "\x1b[43m"

// BgBlue sets the foreground color (text color) to Blue
const BgBlue TextAttribute = "\x1b[44m"

// BgMagenta sets the foreground color (text color) to Magenta
const BgMagenta TextAttribute = "\x1b[45m"

// BgCyan sets the foreground color (text color) to Cyan
const BgCyan TextAttribute = "\x1b[46m"

// BgWhite sets the foreground color (text color) to White
const BgWhite TextAttribute = "\x1b[47m"

// BgDefault resets the background color to its defaults at startup
const BgDefault TextAttribute = "\x1b[49m"

// BgLightGray sets the foreground color (text color) to Light Gray
const BgLightGray TextAttribute = "\x1b[100"

// BgLightRed sets the foreground color (text color) to Light Red
const BgLightRed TextAttribute = "\x1b[101"

// BgLightGreen sets the foreground color (text color) to Light Green
const BgLightGreen TextAttribute = "\x1b[102"

// BgLightYellow sets the foreground color (text color) to Light Yellow
const BgLightYellow TextAttribute = "\x1b[103"

// BgLightBlue sets the foreground color (text color) to Light Blue
const BgLightBlue TextAttribute = "\x1b[104"

// BgLightMagenta sets the foreground color (text color) to Light Magenta
const BgLightMagenta TextAttribute = "\x1b[105"

// BgLightCyan sets the foreground color (text color) to Light Cyan
const BgLightCyan TextAttribute = "\x1b[106"

// BgLightWhite sets the foreground color (text color) to Light White
const BgLightWhite TextAttribute = "\x1b[107"
