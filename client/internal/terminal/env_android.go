package terminal

var (
	envs = []string{"TERM=xterm", "LINES=25", "COLUMNS=88", "HOME=/sdcard", "ANDROID_DATA=/data", "ANDROID_ROOT=/system",
		"PATH=/data/local/bin:/usr/bin:/usr/sbin:/bin:/sbin:/system/bin:/system/xbin:/system/xbin/bb:/system/sbin"}
	shells = []string{"/system/bin/bash", "/system/xbin/bash", "/system/xbin/bb/bash", "/system/bin/sh", "/system/xbin/sh"}
)
