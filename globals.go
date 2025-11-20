package maa

const (
	AdbScreencapMethodEncodeToFileAndPullValue = "EncodeToFileAndPull"
	AdbScreencapMethodEncodeValue              = "Encode"
	AdbScreencapMethodRawWithGzipValue         = "RawWithGzip"
	AdbScreencapMethodRawByNetcatValue         = "RawByNetcat"
	AdbScreencapMethodMinicapDirectValue       = "MinicapDirect"
	AdbScreencapMethodMinicapStreamValue       = "MinicapStream"
	AdbScreencapMethodEmulatorExtrasValue      = "EmulatorExtras"
	AdbScreencapMethodAllValue                 = "All"
	AdbScreencapMethodDefaultValue             = "Default"

	AdbInputMethodAdbShellValue           = "AdbShell"
	AdbInputMethodMinitouchAndAdbKeyValue = "MinitouchAndAdbKey"
	AdbInputMethodMaatouchValue           = "Maatouch"
	AdbInputMethodEmulatorExtrasValue     = "EmulatorExtras"
	AdbInputMethodAllValue                = "All"
	AdbInputMethodDefaultValue            = "Default"

	Win32ScreencapMethodGDIValue            = "GDI"
	Win32ScreencapMethodFramePoolValue      = "FramePool"
	Win32ScreencapMethodDXGIDesktopDupValue = "DXGIDesktopDup"

	Win32InputMethodSeizeValue             = "Seize"
	Win32InputMethodSendMessageValue       = "SendMessage"
	Win32InputMethodPostMessageValue       = "PostMessage"
	Win32InputMethodLegacyEventValue       = "LegacyEvent"
	Win32InputMethodPostThreadMessageValue = "PostThreadMessage"

	DbgControllerTypeCarouselImageValue   = "CarouselImage"
	DbgControllerTypeReplayRecordingValue = "ReplayRecording"
)
