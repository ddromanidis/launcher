package launcher

import (
	"context"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

// ANSI Color Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

var _ Runnable = (*Launcher)(nil)

type Launcher struct {
	ls []Runnable
}

func NewLauncher(ls ...Runnable) Launcher {
	return Launcher{ls: ls}
}

func (l Launcher) Run(ctx context.Context) error {
	g, gCtx := errgroup.WithContext(ctx)

	for _, lnch := range l.ls {
		g.Go(func() error {
			return lnch.Run(gCtx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("an app in launcher has returned an error: %w", err)
	}

	return nil
}

func Launching(ls ...Runnable) Runnable {
	return NewLauncher(ls...)
}

func LaunchingWithAwesomeAnimation(ls ...Runnable) Runnable {
	// --- Countdown Phase ---
	countdownNumbers := map[int]string{
		3: `
     ##### 
         # 
     ##### 
         # 
     ##### 
`,
		2: `
     ##### 
         # 
     ##### 
     #     
     ##### 
`,
		1: `
       #   
      ##   
       #   
       #   
     ##### 
`,
	}

	for i := 3; i >= 1; i-- {
		clearScreen()
		fmt.Println(Red + countdownNumbers[i] + Reset)
		time.Sleep(1 * time.Second)
	}

	clearScreen()
	fmt.Println(Cyan + `
#          #     #     # ##    #  ##### #     # 
#         # #    #     # # #   # #      #     # 
#        #   #   #     # #  #  # #      ####### 
#       #######  #     # #   # # #      #     # 
####### #     #   #####  #    ##  ##### #     # 
` + Reset)
	time.Sleep(1 * time.Second)

	// --- Rocket Launch Animation ---
	rocketFrames := []string{
		// Frame 1: On the ground
		`
      
      
      
      
      
      
       / \
      | o |
      | o |
     /-----\
    /-------\
   |    _    |
   |  /   \  |
  /|  '---'  |\
 / |_________|\
(===============)
`,
		// Frame 2: Ignition
		`
      
      
      
      
      
       / \
      | o |
      | o |
     /-----\
    /-------\
   |    _    |
   |  /   \  |
  /|  '---'  |\
 / |_________|\
(===============)
` + Yellow + `   
 (*******) 
  (*****)  
	 (***)   
` + Reset,
		// Frame 3: Liftoff
		`
      
      
      
       / \
      | o |
      | o |
     /-----\
    /-------\
   |    _    |
   |  /   \  |
  /|  '---'  |\
 / |_________|\
(===============)
` + Yellow + `  
(*********)
 (*******) 
	(*****)  
` + Reset,
		// Frame 4: Ascending
		`
       / \
      | o |
      | o |
     /-----\
    /-------\
   |    _    |
   |  /   \  |
  /|  '---'  |\
 / |_________|\
(===============)
` + Yellow + ` 
   (*********)
    (*******) 
     (*****)  
      (***)   
       (*)    
` + Reset,
		// Frame 5: Higher
		`
       / \
      | o |
      | o |
     /-----\
    /-------\
   |    _    |
` + Yellow + `
   (*********)
    (*******) 
     (*****)  
      (***)   
       (*)    
      
      
` + Reset,
		// Frame 6: Into the clouds
		`
      
` + Yellow + `
(*********)
 (*******) 
  (*****)  
   (***)   
    (*)    
      
      
      
` + Reset + `
~-~-~-~-~-~-~-~
 ~-~-~-~-~-~-~-~
`,
	}

	for _, frame := range rocketFrames {
		clearScreen()

		// --- FIX IS HERE ---
		// Calculate the required padding
		paddingCount := 10 - len(strings.Split(frame, "\n"))

		// Guard against a negative count
		if paddingCount < 0 {
			paddingCount = 0
		}
		// --- END FIX ---

		padding := strings.Repeat("\n", paddingCount)
		fmt.Print(padding)
		fmt.Println(frame)
		time.Sleep(200 * time.Millisecond)
	}

	clearScreen()
	fmt.Println(
		"\n\n\n\n" + Cyan + "      🚀 Mission Successful! Welcome to orbit. 🚀" + Reset + "\n\n\n\n",
	)

	return NewLauncher(ls...)
}

// clearScreen clears the terminal screen and moves the cursor to the top left.
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}
