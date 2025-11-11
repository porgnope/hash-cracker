package main

import (
	"bufio"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func formatNumber(n int64) string {
	str := strconv.FormatInt(n, 10)
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

func formatFloat(f float64) string {
	n := int64(f)
	return formatNumber(n)
}

func defineHashType(hash string) string {
	hash = strings.TrimSpace(hash)
	hash = strings.ToLower(hash)

	hashTypes := map[int]string{
		32:  "MD5",
		40:  "SHA-1",
		64:  "SHA-256",
		128: "SHA-512",
	}

	if hashType, exists := hashTypes[len(hash)]; exists {
		return hashType
	}
	return "Unknown"
}

func hashPassword(password string, hashType string) string {
	switch hashType {
	case "MD5":
		hash := md5.Sum([]byte(password))
		return hex.EncodeToString(hash[:])
	case "SHA-1":
		hash := sha1.Sum([]byte(password))
		return hex.EncodeToString(hash[:])
	case "SHA-256":
		hash := sha256.Sum256([]byte(password))
		return hex.EncodeToString(hash[:])
	case "SHA-512":
		hash := sha512.Sum512([]byte(password))
		return hex.EncodeToString(hash[:])
	default:
		return ""
	}
}

func loadWordlist(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var passwords []string
	scanner := bufio.NewScanner(file)

	const maxCapacity = 10 * 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if line != "" {
			passwords = append(passwords, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return passwords, nil
}

func generateMutations(password string) []string {
	mutations := []string{password}

	mutations = append(mutations, strings.ToLower(password))
	mutations = append(mutations, strings.ToUpper(password))
	mutations = append(mutations, capitalize(password))

	for i := 0; i <= 9; i++ {
		mutations = append(mutations, password+strconv.Itoa(i))
	}
	mutations = append(mutations, password+"123")
	mutations = append(mutations, password+"12")
	mutations = append(mutations, password+"1234")
	mutations = append(mutations, password+"2024")
	mutations = append(mutations, password+"2025")
	mutations = append(mutations, password+"23022025")
	mutations = append(mutations, password+"_23022025")
	mutations = append(mutations, password+"!")
	mutations = append(mutations, password+"!!")
	mutations = append(mutations, password+"@")
	mutations = append(mutations, password+"#")

	leetPassword := leetspeak(password)
	if leetPassword != password {
		mutations = append(mutations, leetPassword)
		mutations = append(mutations, capitalize(leetPassword))
	}

	capPassword := capitalize(password)
	mutations = append(mutations, capPassword+"1")
	mutations = append(mutations, capPassword+"123")
	mutations = append(mutations, capPassword+"!")
	mutations = append(mutations, capPassword+"@")
	mutations = append(mutations, capPassword+"2024")
	mutations = append(mutations, capPassword+"2025")
	mutations = append(mutations, capPassword+"23022025")
	mutations = append(mutations, capPassword+"_23022025")

	mutations = append(mutations, "1"+password)
	mutations = append(mutations, "12"+password)
	mutations = append(mutations, "123"+password)

	return removeDuplicates(mutations)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func leetspeak(s string) string {
	replacer := strings.NewReplacer(
		"a", "@",
		"A", "@",
		"e", "3",
		"E", "3",
		"i", "1",
		"I", "1",
		"o", "0",
		"O", "0",
		"s", "$",
		"S", "$",
		"t", "7",
		"T", "7",
	)
	return replacer.Replace(s)
}

func removeDuplicates(passwords []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, password := range passwords {
		if !seen[password] {
			seen[password] = true
			result = append(result, password)
		}
	}

	return result
}

func worker(id int, passwords []string, target string, hashType string, found chan<- string, attempts *int64, wg *sync.WaitGroup, done *atomic.Bool) {
	defer wg.Done()

	for _, password := range passwords {
		if done.Load() {
			return
		}

		atomic.AddInt64(attempts, 1)
		hashedPassword := hashPassword(password, hashType)

		if hashedPassword == strings.ToLower(target) {
			if !done.Swap(true) {
				found <- password
			}
			return
		}
	}
}

func dictionaryAttack(target string, filePath string, numWorkers int) {
	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("    Dictionary Attack Started")
	fmt.Println("═══════════════════════════════════")

	start := time.Now()

	hashType := defineHashType(target)
	fmt.Printf("Обнаружен тип хеша: %s\n", hashType)

	if hashType == "Unknown" {
		fmt.Println("✗ Ошибка: не удалось определить тип хеша")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("✗ Ошибка: файл не найден: %s\n", filePath)
		return
	}

	fmt.Printf("Используется словарь: %s\n", filePath)
	fmt.Printf("Целевой хеш: %s\n", target)
	fmt.Printf("CPU ядер: %d\n", runtime.NumCPU())
	fmt.Printf("Количество потоков: %d\n", numWorkers)

	fmt.Print("Загрузка словаря... ")
	passwords, err := loadWordlist(filePath)
	if err != nil {
		fmt.Printf("✗ Ошибка загрузки файла: %v\n", err)
		return
	}
	fmt.Printf("OK (%s паролей)\n", formatNumber(int64(len(passwords))))

	if len(passwords) == 0 {
		fmt.Println("✗ Словарь пуст")
		return
	}

	fmt.Println("Начинаем подбор...\n")

	totalPasswords := len(passwords)
	chunkSize := totalPasswords / numWorkers

	if totalPasswords%numWorkers > 0 {
		chunkSize++
	}

	var attempts int64
	var done atomic.Bool
	found := make(chan string, 1)
	var wg sync.WaitGroup

	stopProgress := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !done.Load() {
					currentAttempts := atomic.LoadInt64(&attempts)
					progress := float64(currentAttempts) / float64(len(passwords)) * 100
					elapsed := time.Since(start).Seconds()
					speed := float64(currentAttempts) / elapsed

					fmt.Printf("Проверено: %s/%s паролей (%.1f%%) | Скорость: %s п/сек\n",
						formatNumber(currentAttempts),
						formatNumber(int64(len(passwords))),
						progress,
						formatFloat(speed))
				}
			case <-stopProgress:
				return
			}
		}
	}()

	for i := 0; i < numWorkers; i++ {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize

		if i == numWorkers-1 {
			endIdx = totalPasswords
		}

		if startIdx >= totalPasswords {
			break
		}
		if endIdx > totalPasswords {
			endIdx = totalPasswords
		}

		wg.Add(1)
		go worker(i, passwords[startIdx:endIdx], target, hashType, found, &attempts, &wg, &done)
	}

	go func() {
		wg.Wait()
		if !done.Load() {
			close(found)
		}
	}()

	password, ok := <-found
	close(stopProgress)
	duration := time.Since(start)
	finalAttempts := atomic.LoadInt64(&attempts)

	if ok && password != "" {
		speed := float64(finalAttempts) / duration.Seconds()
		fmt.Println("\n═══════════════════════════════════")
		fmt.Printf("✓ Пароль найден: %s\n", password)
		fmt.Printf("Попыток: %s\n", formatNumber(finalAttempts))
		fmt.Printf("Время выполнения: %v\n", duration)
		if duration.Seconds() > 0 {
			fmt.Printf("Средняя скорость: %s паролей/сек\n", formatFloat(speed))
		}
		fmt.Println("═══════════════════════════════════")
	} else {
		speed := float64(finalAttempts) / duration.Seconds()
		fmt.Println("\n═══════════════════════════════════")
		fmt.Printf("✗ Пароль не найден\n")
		fmt.Printf("Проверено паролей: %s\n", formatNumber(finalAttempts))
		fmt.Printf("Время выполнения: %v\n", duration)
		if duration.Seconds() > 0 {
			fmt.Printf("Средняя скорость: %s паролей/сек\n", formatFloat(speed))
		}
		fmt.Println("═══════════════════════════════════")
	}
}

func dictionaryAttackWithMutations(target string, filePath string, numWorkers int) {
	fmt.Println("\n═══════════════════════════════════")
	fmt.Println("  Dictionary Attack + Mutations")
	fmt.Println("═══════════════════════════════════")

	totalStart := time.Now()

	hashType := defineHashType(target)
	fmt.Printf("Обнаружен тип хеша: %s\n", hashType)

	if hashType == "Unknown" {
		fmt.Println("✗ Ошибка: не удалось определить тип хеша")
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("✗ Ошибка: файл не найден: %s\n", filePath)
		return
	}

	fmt.Printf("Используется словарь: %s\n", filePath)
	fmt.Printf("Целевой хеш: %s\n", target)
	fmt.Printf("CPU ядер: %d\n", runtime.NumCPU())
	fmt.Printf("Количество потоков: %d\n", numWorkers)

	// Загрузка словаря
	loadStart := time.Now()
	fmt.Print("Загрузка словаря... ")
	basePasswords, err := loadWordlist(filePath)
	if err != nil {
		fmt.Printf("✗ Ошибка загрузки файла: %v\n", err)
		return
	}
	loadDuration := time.Since(loadStart)
	fmt.Printf("OK (%s базовых паролей)\n", formatNumber(int64(len(basePasswords))))

	if len(basePasswords) == 0 {
		fmt.Println("✗ Словарь пуст")
		return
	}

	// Параллельная генерация мутаций
	mutationStart := time.Now()
	fmt.Print("Генерация мутаций... ")

	// Разделяем работу между горутинами
	chunkSize := len(basePasswords) / numWorkers
	if chunkSize == 0 {
		chunkSize = 1
	}

	type MutationResult struct {
		mutations []string
	}

	resultChan := make(chan MutationResult, numWorkers)
	var mutWg sync.WaitGroup

	// Запускаем горутины для генерации мутаций
	for i := 0; i < numWorkers; i++ {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize

		if i == numWorkers-1 {
			endIdx = len(basePasswords)
		}

		if startIdx >= len(basePasswords) {
			break
		}
		if endIdx > len(basePasswords) {
			endIdx = len(basePasswords)
		}

		mutWg.Add(1)
		go func(passwords []string) {
			defer mutWg.Done()
			var localMutations []string
			for _, basePassword := range passwords {
				mutations := generateMutations(basePassword)
				localMutations = append(localMutations, mutations...)
			}
			resultChan <- MutationResult{mutations: localMutations}
		}(basePasswords[startIdx:endIdx])
	}

	// Собираем результаты
	go func() {
		mutWg.Wait()
		close(resultChan)
	}()

	var allPasswords []string
	for result := range resultChan {
		allPasswords = append(allPasswords, result.mutations...)
	}

	mutationDuration := time.Since(mutationStart)
	fmt.Printf("OK (%s вариантов)\n", formatNumber(int64(len(allPasswords))))
	fmt.Printf("Увеличение: x%.1f\n", float64(len(allPasswords))/float64(len(basePasswords)))
	fmt.Printf("Время генерации мутаций: %v\n", mutationDuration)

	fmt.Println("Начинаем подбор...\n")

	// Начало хеширования
	hashingStart := time.Now()

	totalPasswords := len(allPasswords)
	chunkSize = totalPasswords / numWorkers

	if totalPasswords%numWorkers > 0 {
		chunkSize++
	}

	var attempts int64
	var done atomic.Bool
	found := make(chan string, 1)
	var wg sync.WaitGroup

	stopProgress := make(chan bool)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !done.Load() {
					currentAttempts := atomic.LoadInt64(&attempts)
					progress := float64(currentAttempts) / float64(len(allPasswords)) * 100
					elapsed := time.Since(hashingStart).Seconds()
					speed := float64(currentAttempts) / elapsed

					fmt.Printf("Проверено: %s/%s паролей (%.1f%%) | Скорость: %s п/сек\n",
						formatNumber(currentAttempts),
						formatNumber(int64(len(allPasswords))),
						progress,
						formatFloat(speed))
				}
			case <-stopProgress:
				return
			}
		}
	}()

	for i := 0; i < numWorkers; i++ {
		startIdx := i * chunkSize
		endIdx := startIdx + chunkSize

		if i == numWorkers-1 {
			endIdx = totalPasswords
		}

		if startIdx >= totalPasswords {
			break
		}
		if endIdx > totalPasswords {
			endIdx = totalPasswords
		}

		wg.Add(1)
		go worker(i, allPasswords[startIdx:endIdx], target, hashType, found, &attempts, &wg, &done)
	}

	go func() {
		wg.Wait()
		if !done.Load() {
			close(found)
		}
	}()

	password, ok := <-found
	close(stopProgress)

	// Замеряем время
	hashingDuration := time.Since(hashingStart)
	totalDuration := time.Since(totalStart)
	finalAttempts := atomic.LoadInt64(&attempts)

	if ok && password != "" {
		speed := float64(finalAttempts) / hashingDuration.Seconds()
		fmt.Println("\n═══════════════════════════════════")
		fmt.Printf("✓ Пароль найден: %s\n", password)
		fmt.Printf("Попыток: %s\n", formatNumber(finalAttempts))
		fmt.Println("\n--- Статистика производительности ---")
		fmt.Printf("Время загрузки словаря: %v\n", loadDuration)
		fmt.Printf("Время генерации мутаций: %v \n", mutationDuration)
		fmt.Printf("Время хеширования: %v\n", hashingDuration)
		fmt.Printf("Общее время выполнения: %v\n", totalDuration)
		if hashingDuration.Seconds() > 0 {
			fmt.Printf("Скорость хеширования: %s паролей/сек\n", formatFloat(speed))
		}
		fmt.Println("═══════════════════════════════════")
	} else {
		speed := float64(finalAttempts) / hashingDuration.Seconds()
		fmt.Println("\n═══════════════════════════════════")
		fmt.Printf("✗ Пароль не найден\n")
		fmt.Printf("Проверено паролей: %s\n", formatNumber(finalAttempts))
		fmt.Println("\n--- Статистика производительности ---")
		fmt.Printf("Время загрузки словаря: %v\n", loadDuration)
		fmt.Printf("Время генерации мутаций: %v \n", mutationDuration)
		fmt.Printf("Время хеширования: %v\n", hashingDuration)
		fmt.Printf("Общее время выполнения: %v\n", totalDuration)
		if hashingDuration.Seconds() > 0 {
			fmt.Printf("Скорость хеширования: %s паролей/сек\n", formatFloat(speed))
		}
		fmt.Println("═══════════════════════════════════")
	}
}
