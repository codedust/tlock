package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func main() {
	// Config
	var password string = "secretpassword"
	var bgcolors bool = true
	var jail bool = false
	var scanner bool = true
	var cross bool = false
	var faster bool = false
	var autostart bool = false

	// Init
	var iteration int = -1
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	reader := bufio.NewReader(os.Stdin)

	defer fmt.Print("\033[?25h") // show cursor again when we exit
	defer fmt.Print("\033c")     // clear the screen (nocolor) when we exit

	// show fake su environment
	hostname, _ := os.Hostname()
	fmt.Print("\033c\033[H[root@" + hostname + " ~]# cd /\n")
	fmt.Print("[root@" + hostname + " /]# ls\n")
	fmt.Print("bin  boot  dev  etc  home  lib  lib64  mnt  opt  proc  root  run  sbin  srv  sys  tmp  usr  var\n")
	fmt.Print("[root@" + hostname + " /]# cd /etc/sudoers\n")
	fmt.Print("-bash: cd: /etc/sudoers: No such file or directory\n")
	fmt.Print("[root@" + hostname + " /]# vim /etc/sudoers\n")
	fmt.Print("[root@" + hostname + " /]# ")

	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run() // disable stdin buffering
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()              // do not echo input
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()         // reenable echoing on exit

	cols, rows, _ := GetTerminalSize()

	var amplifier int
	if faster {
		amplifier = 10
	} else {
		amplifier = 1
	}

	// Animation
	go func() {
		for {
			if autostart {
				fmt.Print("\033[?25l") // hide cursor

				iteration++

				// draw different animations
				if jail {
					if bgcolors {
						fmt.Print("\033[4" + strconv.Itoa(iteration*amplifier%7) + "m") // set background color
					} else {
						fmt.Print("\033[40m") // set background color
					}

					fmt.Print("\033[2J") // clear screen (current bgcolor)

					fmt.Print("\033[4" + strconv.Itoa(iteration*amplifier%7+1) + "m") // set background color
					for y := 1; y <= rows; y++ {
						for x := 2; x <= cols; x += 5 {
							fmt.Print("\033[" + strconv.Itoa(y) + ";" + strconv.Itoa(x) + "f ")
						}
						time.Sleep(10 * time.Millisecond)
					}

				} else {
					if bgcolors {
						fmt.Print("\033[4" + strconv.Itoa(iteration*amplifier/rows%7) + "m") // set background color
					} else {
						fmt.Print("\033[40m") // set background color
					}

					fmt.Print("\033[2J") // clear screen (current bgcolor)
				}
				if scanner || cross {
					fmt.Print("\033[4" + strconv.Itoa(iteration*amplifier/rows%7+1) + "m") // set background color
					y := (iteration%rows + 1)
					if iteration%(rows*2) >= rows {
						y = rows - y
					}
					for x := 1; x <= cols; x++ {
						fmt.Print("\033[" + strconv.Itoa(y) + ";" + strconv.Itoa(x) + "f ")
					}
				}

				if cross {
					x := (iteration%cols + 1)
					if iteration%(cols*2) >= cols {
						x = cols - x
					}
					for y := 1; y <= rows; y++ {
						fmt.Print("\033[" + strconv.Itoa(y) + ";" + strconv.Itoa(x) + "f ")
					}
				}
			}

			time.Sleep(50 * time.Millisecond)
		}
	}()

	// Password check
	inputstring := ""

	for {
		input, _ := reader.ReadByte()
		inputstring += string(input)

		autostart = true

		if strings.Contains(inputstring, password) {
			return
		}
	}
}

// GetTerminalSize returns the cols and rows of the terminal
func GetTerminalSize() (int, int, error) {
	out, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0)
	if err != nil {
		return 0, 0, err
	}
	defer out.Close()

	// fd is the integer Unix file descriptor referencing the open file
	var fd uintptr = out.Fd()

	type winsize struct {
		rows    uint16
		cols    uint16
		xpixels uint16
		ypixels uint16
	}

	var sz winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))

	return int(sz.cols), int(sz.rows), nil
}
