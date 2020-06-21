package main
import ("fmt"
"bufio"
"os")
func main(){
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("guess a no. between 1 to 100")
	fmt.Println("  press ENTER when ready")
	scanner.Scan()
	upperLMT:=100
	guess:=50
	lowerLMT:=1
	response := scanner.Text()
	for{
		fmt.Println("i guess the no. is",guess)
		fmt.Println("is that:")
		fmt.Println("(a) high compared to your guess")
		fmt.Println("(b) low compared to your guess")
		fmt.Println("(c) correct")
		scanner.Scan()
		response = scanner.Text()
		if response == "a" {
			upperLMT=guess
			guess=int(float64(upperLMT + lowerLMT)*0.5)
		} else if response == "b" {
			lowerLMT=guess
			guess=int(float64(upperLMT + lowerLMT)*0.5)
		} else if response == "c"{
			upperLMT=100
			guess=50
			lowerLMT=1
			fmt.Println("i WON!")
			fmt.Println("  press ENTER to play again")
			scanner.Scan()
			response = scanner.Text()
			if response == "exit" {break}
			fmt.Println()
			fmt.Println("guess a no. between 1 to 100")
			fmt.Println("  press ENTER when ready")
			scanner.Scan()
			response = scanner.Text()
			if response == "exit" {break}
		}else if response == "exit" {
			break
		} else{
			fmt.Println("invalid Response! try again")
		}
	}
}