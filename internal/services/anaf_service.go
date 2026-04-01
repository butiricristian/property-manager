package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ANAFService struct {
	client *http.Client
}

func NewANAFService() *ANAFService {
	return &ANAFService{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type ANAFCompany struct {
	Name            string `json:"name"`
	CUI             string `json:"cui"`
	Street          string `json:"street"`
	City            string `json:"city"`
	County          string `json:"county"`
	PostalCode      string `json:"postal_code"`
	TradeRegisterNr string `json:"trade_register_nr"`
	VATRegistered   bool   `json:"vat_registered"`
}

type anafRequest struct {
	CUI  int64  `json:"cui"`
	Data string `json:"data"`
}

type anafResponse struct {
	Found []anafFound `json:"found"`
}

type anafFound struct {
	DateGenerale anafDateGenerale `json:"date_generale"`
	InregistrareScopTVA anafScopTVA `json:"inregistrare_scop_Tva"`
}

type anafDateGenerale struct {
	Denumire    string `json:"denumire"`
	Adresa      string `json:"adresa"`
	NrRegCom    string `json:"nrRegCom"`
	CodPostal   string `json:"cod_postal"`
	StatusRO_e_Factura bool `json:"statusRO_e_Factura"`
}

type anafScopTVA struct {
	ScpTVA bool `json:"scpTVA"`
}

func (s *ANAFService) LookupCompany(ctx context.Context, cui string) (*ANAFCompany, error) {
	// Strip "RO" prefix if present
	cuiClean := strings.TrimPrefix(strings.ToUpper(cui), "RO")
	cuiClean = strings.TrimSpace(cuiClean)

	var cuiNum int64
	if _, err := fmt.Sscanf(cuiClean, "%d", &cuiNum); err != nil {
		return nil, fmt.Errorf("invalid CUI: %s", cui)
	}

	today := time.Now().Format("2006-01-02")
	body, err := json.Marshal([]anafRequest{{CUI: cuiNum, Data: today}})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://webservicesp.anaf.ro/api/PlatitorTvaRest/v9/tva",
		bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ANAF API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ANAF API returned status %d", resp.StatusCode)
	}

	var result anafResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ANAF response: %w", err)
	}

	if len(result.Found) == 0 {
		return nil, fmt.Errorf("no company found for CUI %s", cui)
	}

	found := result.Found[0]
	company := &ANAFCompany{
		Name:            found.DateGenerale.Denumire,
		CUI:             cuiClean,
		TradeRegisterNr: found.DateGenerale.NrRegCom,
		PostalCode:      found.DateGenerale.CodPostal,
		VATRegistered:   found.InregistrareScopTVA.ScpTVA,
	}

	// Parse address into components
	parseAddress(found.DateGenerale.Adresa, company)

	return company, nil
}

// countyNameToISO maps Romanian county names (as returned by ANAF) to ISO 3166-2:RO codes.
var countyNameToISO = map[string]string{
	"ALBA":            "AB",
	"ARAD":            "AR",
	"ARGES":           "AG",
	"BACAU":           "BC",
	"BIHOR":           "BH",
	"BISTRITA-NASAUD": "BN",
	"BISTRITA NASAUD": "BN",
	"BOTOSANI":        "BT",
	"BRAILA":          "BR",
	"BRASOV":          "BV",
	"BUCURESTI":       "B",
	"BUZAU":           "BZ",
	"CALARASI":        "CL",
	"CARAS-SEVERIN":   "CS",
	"CARAS SEVERIN":   "CS",
	"CLUJ":            "CJ",
	"CONSTANTA":       "CT",
	"COVASNA":         "CV",
	"DAMBOVITA":       "DB",
	"DOLJ":            "DJ",
	"GALATI":          "GL",
	"GIURGIU":         "GR",
	"GORJ":            "GJ",
	"HARGHITA":        "HR",
	"HUNEDOARA":       "HD",
	"IALOMITA":        "IL",
	"IASI":            "IS",
	"ILFOV":           "IF",
	"MARAMURES":       "MM",
	"MEHEDINTI":       "MH",
	"MURES":           "MS",
	"NEAMT":           "NT",
	"OLT":             "OT",
	"PRAHOVA":         "PH",
	"SALAJ":           "SJ",
	"SATU MARE":       "SM",
	"SATU-MARE":       "SM",
	"SIBIU":           "SB",
	"SUCEAVA":         "SV",
	"TELEORMAN":       "TR",
	"TIMIS":           "TM",
	"TULCEA":          "TL",
	"VALCEA":          "VL",
	"VASLUI":          "VS",
	"VRANCEA":         "VN",
}

func mapCountyToISO(name string) string {
	name = strings.ToUpper(strings.TrimSpace(name))
	if code, ok := countyNameToISO[name]; ok {
		return code
	}
	// If it's already a short code (1-2 chars), return as-is
	if len(name) <= 2 {
		return name
	}
	return name
}

func parseAddress(addr string, c *ANAFCompany) {
	// ANAF returns addresses like: "Jud.ALBA, Loc.ALBA IULIA, Str.Strada X, Nr.10"
	parts := strings.Split(addr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lower := strings.ToLower(part)
		switch {
		case strings.HasPrefix(lower, "jud."):
			countyName := strings.TrimSpace(part[4:])
			c.County = mapCountyToISO(countyName)
		case strings.HasPrefix(lower, "mun.") || strings.HasPrefix(lower, "loc.") || strings.HasPrefix(lower, "oras."):
			idx := strings.Index(part, ".")
			c.City = strings.TrimSpace(part[idx+1:])
		case strings.HasPrefix(lower, "str."):
			c.Street = strings.TrimSpace(part[4:])
		default:
			// Append extra address parts to street
			if c.Street != "" && part != "" {
				c.Street += ", " + part
			}
		}
	}

	// If no county was parsed but city is BUCURESTI, set county to B
	if c.County == "" && strings.Contains(strings.ToUpper(c.City), "BUCURESTI") {
		c.County = "B"
	}

	// For Bucharest, map city to SECTOR format
	// ANAF may return "BUCURESTI SECTORUL 3" or "Sector 3" etc.
	if c.County == "B" {
		city := strings.ToUpper(c.City)
		for i := 1; i <= 6; i++ {
			s := fmt.Sprintf("%d", i)
			if strings.Contains(city, "SECTOR") && strings.Contains(city, s) {
				c.City = fmt.Sprintf("SECTOR%d", i)
				break
			}
		}
	}
}
