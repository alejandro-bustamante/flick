package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alejandro-bustamante/flick/internal/config"
	"github.com/alejandro-bustamante/flick/internal/core/parser"
	"github.com/alejandro-bustamante/flick/internal/utils"
	"github.com/gocolly/colly/v2"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(scrapeCmd)
}

const (
	baseURL      = "https://www.limetorrents.fun"
	popularURL   = baseURL + "/browse-torrents/Movies/date"
	outputFile   = "scraper_results.txt"
	torrentsHost = "https://www.limetorrents.fun"
	maxPages     = 10 // Número máximo de páginas a scrapear
)

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Scrapes limetorrents for new junk patterns.",
	Long: `Scrapes the popular movies page from limetorrents.fun,
extracts torrent filenames, and uses the internal parser to
find potential new junk patterns not present in patterns.toml.
Results are saved to scraper_results.txt for manual review.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting scraper...")

		// --- 1. Cargar configuración y parser existentes ---
		_, patternsPath, err := config.GetConfigPaths(cfgFile)
		if err != nil {
			log.Fatalf("Error getting config paths: %v", err)
		}

		data, err := config.LoadPatterns(patternsPath)
		if err != nil {
			log.Fatalf("Error loading patterns.toml from %s: %v", patternsPath, err)
		}

		logger := utils.NewLogger("debug")
		p := parser.NewMediaParser(
			data.Tokenizer.Separators,
			data.Cleaner.JunkPatterns,
			data.Extractor.YearRange[:],
			logger,
		)

		existingPatterns := make(map[string]struct{})
		for _, pattern := range data.Cleaner.JunkPatterns {
			existingPatterns[strings.ToLower(pattern)] = struct{}{}
		}
		log.Printf("Loaded %d existing junk patterns.", len(existingPatterns))

		// --- 2. Preparar archivo de salida ---
		file, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("Failed to create output file %s: %v", outputFile, err)
		}
		defer file.Close()
		log.Printf("Output will be saved to %s", outputFile)

		var newPatternsFound int
		var movieLinks []string

		// --- 3. Configurar Colly ---
		c := colly.NewCollector(
			colly.UserAgent("Flick-Scraper-Bot/1.0 (compatible; +https://github.com/alejandro-bustamante/flick)"),
		)

		detailCollector := c.Clone()

		// --- 4. Definir Callbacks de Colly ---

		// Recolectar todos los enlaces de películas
		c.OnHTML("div.tt-name a:nth-child(2)", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if link == "" {
				return
			}
			absoluteURL := e.Request.AbsoluteURL(link)
			movieLinks = append(movieLinks, absoluteURL)
			log.Printf("Found movie link: %s", absoluteURL)
		})

		detailCollector.OnHTML("div.fileline span.csprite_doc_video", func(e *colly.HTMLElement) {
			// Obtener el HTML completo del elemento padre
			parentHTML, err := e.DOM.Parent().Html()
			if err != nil {
				log.Printf("WARN: Failed to get parent HTML on %s: %v", e.Request.URL, err)
				return
			}

			// Encontrar la posición del span con clase csprite_doc_video
			videoSpanMarker := `<span class="csprite_doc_video"></span>`
			idx := strings.Index(parentHTML, videoSpanMarker)
			if idx == -1 {
				log.Printf("WARN: Could not find video span marker in HTML on %s", e.Request.URL)
				return
			}

			// Extraer el texto después del span
			textAfterSpan := parentHTML[idx+len(videoSpanMarker):]

			// Remover cualquier tag HTML restante y obtener solo texto
			textAfterSpan = strings.TrimSpace(textAfterSpan)

			// Buscar el nombre del archivo (termina en .mkv o .mp4 antes del guion y tamaño)
			// Ejemplo: " The 12th Man 2017 1080p BluRay x264-nikt0.mkv -  <div..."
			var filename string

			// Dividir por tags HTML para limpiar
			parts := strings.Split(textAfterSpan, "<")
			if len(parts) > 0 {
				firstPart := strings.TrimSpace(parts[0])

				// Ahora dividir por " - " para separar el nombre del archivo del tamaño
				fileParts := strings.Split(firstPart, " - ")
				if len(fileParts) > 0 {
					filename = strings.TrimSpace(fileParts[0])
				}
			}

			if filename == "" {
				log.Printf("WARN: No filename found after video span on %s", e.Request.URL)
				return
			}

			// --- 5. Usar el parser ---
			result := p.ParseNormalized(filename)
			if result.MediaInfo == nil {
				log.Printf("WARN: Parser failed for: %s", filename)
				return
			}

			potentialJunk := result.MediaInfo.Remaining
			if len(potentialJunk) == 0 {
				return
			}

			// --- 6. Filtrar patrones ---
			var newJunk []string
			for _, token := range potentialJunk {
				lowerToken := strings.ToLower(token)
				if _, exists := existingPatterns[lowerToken]; !exists {
					newJunk = append(newJunk, token)
					existingPatterns[lowerToken] = struct{}{}
				}
			}

			// --- 7. Escribir en el archivo ---
			if len(newJunk) > 0 {
				outputLine := fmt.Sprintf("File: %s | New Patterns: %v\n", filename, newJunk)
				if _, err := file.WriteString(outputLine); err != nil {
					log.Printf("WARN: Failed to write to output file: %v", err)
				}
				newPatternsFound += len(newJunk)
			}
		})

		// Callbacks de logging y errores
		c.OnRequest(func(r *colly.Request) {
			log.Println("Fetching popular movies page:", r.URL)
		})

		detailCollector.OnRequest(func(r *colly.Request) {
			log.Println("Scraping movie page:", r.URL)
		})

		c.OnError(func(r *colly.Response, err error) {
			log.Printf("ERROR: Main collector failed for %s: %v", r.Request.URL, err)
		})

		detailCollector.OnError(func(r *colly.Response, err error) {
			log.Printf("ERROR: Detail collector failed for %s: %v", r.Request.URL, err)
		})

		// --- 5. Iniciar el scraping ---

		// Primero, recolectar todos los enlaces de todas las páginas
		log.Printf("Collecting movie links from %d pages...", maxPages)
		for page := 1; page <= maxPages; page++ {
			pageURL := fmt.Sprintf("%s/%d", popularURL, page)
			log.Printf("Visiting page %d: %s", page, pageURL)
			if err := c.Visit(pageURL); err != nil {
				log.Printf("WARN: Failed to visit page %d: %v", page, err)
			}
		}

		c.Wait()
		log.Printf("Collected %d movie links. Now scraping details...", len(movieLinks))

		// Ahora visitar cada enlace de película
		for i, link := range movieLinks {
			log.Printf("Scraping movie %d/%d: %s", i+1, len(movieLinks), link)
			if err := detailCollector.Visit(link); err != nil {
				log.Printf("WARN: Failed to visit %s: %v", link, err)
			}
		}

		detailCollector.Wait()

		log.Printf("Scraping finished. Found %d potential new patterns.", newPatternsFound)
		log.Printf("Please review %s manually.", outputFile)
	},
}
