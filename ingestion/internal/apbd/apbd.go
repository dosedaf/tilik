package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type TargetConfig struct {
	Years      []string
	Periods    []string
	Provinces  []string
	Pemdas     []string
	MaxWorkers int
	OutputDir  string
}

type DownloadTask struct {
	Year      string
	Period    string
	ProvCode  string
	ProvName  string
	PemdaCode string
	PemdaName string
}

type TaskResult struct {
	Task       DownloadTask
	FilePath   string
	BytesSaved int64
	Error      error
}

const BasePortalURL = "https://djpk.kemenkeu.go.id/portal"

func apbd() {
	config := TargetConfig{
		Years:      []string{"2026"},
		Periods:    []string{"8"},
		Provinces:  []string{"12"},
		Pemdas:     []string{},
		MaxWorkers: 5,
		OutputDir:  "./apbd_downloads",
	}

	fmt.Println("[INFO] =========================================")
	fmt.Println("[INFO] Starting APBD Downloader Pipeline")
	fmt.Printf("[INFO] Output Directory: %s\n", config.OutputDir)
	fmt.Printf("[INFO] Max Workers: %d\n", config.MaxWorkers)
	fmt.Println("[INFO] =========================================")

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		fmt.Printf("[FATAL] Failed to create output directory: %v\n", err)
		panic(err)
	}

	fmt.Println("[INFO] Building download tasks based on configuration...")
	tasks, err := buildDownloadTasks(httpClient, config)
	if err != nil {
		fmt.Printf("[FATAL] Error building tasks: %v\n", err)
		panic(err)
	}

	if len(tasks) == 0 {
		fmt.Println("[WARN] No tasks generated based on current configuration. Exiting.")
		return
	}

	fmt.Printf("[INFO] Generated %d download tasks.\n", len(tasks))
	fmt.Println("[INFO] Spinning up worker pool...")

	tasksChan := make(chan DownloadTask, len(tasks))
	resultsChan := make(chan TaskResult, len(tasks))
	var wg sync.WaitGroup

	for i := 1; i <= config.MaxWorkers; i++ {
		wg.Add(1)
		go worker(i, httpClient, config.OutputDir, tasksChan, resultsChan, &wg)
	}

	fmt.Println("[INFO] Enqueueing tasks...")
	for _, task := range tasks {
		tasksChan <- task
	}
	close(tasksChan)
	fmt.Println("[INFO] All tasks enqueued. Waiting for workers to finish...")

	go func() {
		wg.Wait()
		close(resultsChan)
		fmt.Println("[INFO] All workers have finished.")
	}()

	var successCount, failureCount int
	for res := range resultsChan {
		if res.Error != nil {
			fmt.Printf("[FAILED] Year:%s Period:%s Prov:%s Pemda:%s | Error: %v\n",
				res.Task.Year, res.Task.Period, res.Task.ProvName, res.Task.PemdaName, res.Error)
			failureCount++
		} else {
			fmt.Printf("[SUCCESS] Saved %s (%d bytes)\n", res.FilePath, res.BytesSaved)
			successCount++
		}
	}

	fmt.Println("[INFO] =========================================")
	fmt.Printf("[INFO] Pipeline Summary: %d Succeeded, %d Failed\n", successCount, failureCount)
	fmt.Println("[INFO] =========================================")
}

func buildDownloadTasks(client *http.Client, config TargetConfig) ([]DownloadTask, error) {
	var tasks []DownloadTask

	years := config.Years
	if len(years) == 0 {
		years = []string{"2026"}
	}

	for _, year := range years {
		fmt.Printf("[DEBUG] Fetching available provinces for year %s...\n", year)
		provMap, err := fetchProvinces(client, year)
		if err != nil {
			return nil, fmt.Errorf("failed fetching provinces for year %s: %w", year, err)
		}
		fmt.Printf("[DEBUG] Found %d provinces.\n", len(provMap)-1)

		targetProvinces := config.Provinces
		if len(targetProvinces) == 0 {
			for provCode := range provMap {
				if provCode != "--" {
					targetProvinces = append(targetProvinces, provCode)
				}
			}
		}

		for _, provCode := range targetProvinces {
			provName := provMap[provCode]
			
			fmt.Printf("[DEBUG] Fetching available pemdas for %s in year %s...\n", provName, year)
			pemdaMap, err := fetchPemdas(client, provCode, year)
			if err != nil {
				return nil, fmt.Errorf("failed fetching pemdas for prov %s year %s: %w", provCode, year, err)
			}
			fmt.Printf("[DEBUG] Found %d pemdas for %s.\n", len(pemdaMap)-1, provName)

			targetPemdas := config.Pemdas
			if len(targetPemdas) == 0 {
				for pemdaCode := range pemdaMap {
					if pemdaCode != "--" {
						targetPemdas = append(targetPemdas, pemdaCode)
					}
				}
			}

			periods := config.Periods
			if len(periods) == 0 {
				periods = []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
			}

			for _, pemdaCode := range targetPemdas {
				pemdaName := pemdaMap[pemdaCode]
				
				for _, period := range periods {
					tasks = append(tasks, DownloadTask{
						Year:      year,
						Period:    period,
						ProvCode:  provCode,
						ProvName:  provName,
						PemdaCode: pemdaCode,
						PemdaName: pemdaName,
					})
				}
			}
		}
	}

	return tasks, nil
}

func fetchProvinces(client *http.Client, year string) (map[string]string, error) {
	reqURL := fmt.Sprintf("%s/provinsi/%s", BasePortalURL, year)
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var provMap map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&provMap); err != nil {
		return nil, err
	}
	return provMap, nil
}

func fetchPemdas(client *http.Client, provCode, year string) (map[string]string, error) {
	reqURL := fmt.Sprintf("%s/pemda/%s/%s", BasePortalURL, provCode, year)
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var pemdaMap map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&pemdaMap); err != nil {
		return nil, err
	}
	return pemdaMap, nil
}

func worker(workerID int, client *http.Client, outputDir string, tasks <-chan DownloadTask, results chan<- TaskResult, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("[WORKER-%d] Started.\n", workerID)

	for task := range tasks {
		fmt.Printf("[WORKER-%d] Processing Task -> Year:%s Period:%s Prov:%s Pemda:%s\n",
			workerID, task.Year, task.Period, task.ProvName, task.PemdaName)

		filePath, bytesSaved, err := executeDownload(client, outputDir, task)
		results <- TaskResult{
			Task:       task,
			FilePath:   filePath,
			BytesSaved: bytesSaved,
			Error:      err,
		}
	}
	fmt.Printf("[WORKER-%d] Shutting down.\n", workerID)
}

func executeDownload(client *http.Client, outputDir string, task DownloadTask) (string, int64, error) {
	params := url.Values{}
	params.Add("type", "apbd")
	params.Add("periode", task.Period)
	params.Add("tahun", task.Year)
	params.Add("provinsi", task.ProvCode)
	params.Add("pemda", task.PemdaCode)

	downloadURL := fmt.Sprintf("%s/csv_apbd?%s", BasePortalURL, params.Encode())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("http status %d", resp.StatusCode)
	}

	cleanProv := strings.ReplaceAll(task.ProvName, " ", "_")
	cleanPemda := strings.ReplaceAll(task.PemdaName, " ", "_")
	
	filename := fmt.Sprintf("apbd_%s_p%s_%s_%s.xls", task.Year, task.Period, cleanProv, cleanPemda)
	fullPath := filepath.Join(outputDir, filename)

	file, err := os.Create(fullPath)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	bytesSaved, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", 0, err
	}

	return fullPath, bytesSaved, nil
}
