package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func showMenu() {
	numCPU := runtime.NumCPU()
	physicalCores := numCPU / 2

	fmt.Println("\n╔════════════════════════════════════╗")
	fmt.Println("║      Hash Cracker v1.0             ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Printf("Логических ядер (потоков): %d\n", numCPU)
	fmt.Printf("Предполагаемые физические ядра: ~%d\n", physicalCores)
	fmt.Printf("Рекомендуется для хеширования: %d потоков\n", numCPU)
	fmt.Println("\nДоступные методы:")
	fmt.Println("  1 - Dictionary Attack")
	fmt.Println("  2 - Dictionary Attack + Mutations")
	fmt.Println("  0 - Выход")
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		showMenu()

		fmt.Print("\nВведите метод: ")
		methodInput, _ := reader.ReadString('\n')
		methodInput = strings.TrimSpace(methodInput)

		switch methodInput {
		case "1":
			fmt.Print("Введите таргет (hash): ")
			target, _ := reader.ReadString('\n')
			target = strings.TrimSpace(target)

			if target == "" {
				fmt.Println("✗ Ошибка: таргет не может быть пустым")
				continue
			}

			fmt.Print("Введите словарь (файл): ")
			wordlist, _ := reader.ReadString('\n')
			wordlist = strings.TrimSpace(wordlist)

			if wordlist == "" {
				fmt.Println("✗ Ошибка: укажите файл словаря")
				continue
			}

			defaultThreads := runtime.NumCPU()
			fmt.Printf("Введите потоки (автоматически = %d, нажмите Enter): ", defaultThreads)
			threadsInput, _ := reader.ReadString('\n')
			threadsInput = strings.TrimSpace(threadsInput)

			threads := defaultThreads
			if threadsInput != "" {
				if t, err := strconv.Atoi(threadsInput); err == nil && t > 0 {
					threads = t
					if threads > defaultThreads*2 {
						fmt.Printf("⚠ Внимание: %d потоков может быть избыточно для вашего CPU\n", threads)
					}
				} else {
					fmt.Println("⚠ Неверное значение, используется автоматическое определение")
					threads = defaultThreads
				}
			} else {
				fmt.Printf("✓ Автоматически выбрано: %d потоков\n", threads)
			}

			dictionaryAttack(target, wordlist, threads)

		case "2":
			fmt.Print("Введите таргет (hash): ")
			target, _ := reader.ReadString('\n')
			target = strings.TrimSpace(target)

			if target == "" {
				fmt.Println("✗ Ошибка: таргет не может быть пустым")
				continue
			}

			fmt.Print("Введите словарь (файл): ")
			wordlist, _ := reader.ReadString('\n')
			wordlist = strings.TrimSpace(wordlist)

			if wordlist == "" {
				fmt.Println("✗ Ошибка: укажите файл словаря")
				continue
			}

			defaultThreads := runtime.NumCPU()
			fmt.Printf("Введите потоки (автоматически = %d, нажмите Enter): ", defaultThreads)
			threadsInput, _ := reader.ReadString('\n')
			threadsInput = strings.TrimSpace(threadsInput)

			threads := defaultThreads
			if threadsInput != "" {
				if t, err := strconv.Atoi(threadsInput); err == nil && t > 0 {
					threads = t
					if threads > defaultThreads*2 {
						fmt.Printf("⚠ Внимание: %d потоков может быть избыточно для вашего CPU\n", threads)
					}
				} else {
					fmt.Println("⚠ Неверное значение, используется автоматическое определение")
					threads = defaultThreads
				}
			} else {
				fmt.Printf("✓ Автоматически выбрано: %d потоков\n", threads)
			}

			dictionaryAttackWithMutations(target, wordlist, threads)

		case "0":
			fmt.Println("\nВыход из программы...")
			return

		default:
			fmt.Println("\n✗ Неверный выбор. Попробуйте снова.")
		}

		fmt.Print("\nНажмите Enter для продолжения...")
		reader.ReadString('\n')
	}
}
