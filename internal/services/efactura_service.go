package services

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

type InvoiceData struct {
	InvoiceNumber int32
	InvoiceSeries string
	IssueDate     time.Time
	DueDate       time.Time

	SellerName            string
	SellerCUI             string
	SellerCNP             string
	SellerTradeRegisterNr string
	SellerStreet          string
	SellerCity            string
	SellerCounty          string
	SellerPostalCode      string
	SellerCountryCode     string
	SellerVATRegistered   bool

	BuyerName            string
	BuyerCUI             string
	BuyerTradeRegisterNr string
	BuyerStreet          string
	BuyerCity            string
	BuyerCounty          string
	BuyerPostalCode      string
	BuyerCountryCode     string
	BuyerVATRegistered   bool

	Currency    string
	NetAmount   float64
	TaxAmount   float64
	GrossAmount float64
	TaxPercent  float64

	Description string
	Nights      int32
}

func GenerateEFacturaXML(data *InvoiceData) ([]byte, error) {
	inv := ublInvoice{
		XMLNs:    "urn:oasis:names:specification:ubl:schema:xsd:Invoice-2",
		XMLNsCac: "urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2",
		XMLNsCbc: "urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2",

		CustomizationID: "urn:cen.eu:en16931:2017#compliant#urn:efactura.mfinante.ro:CIUS-RO:1.0.1",
		ID:              fmt.Sprintf("%s%d", data.InvoiceSeries, data.InvoiceNumber),
		IssueDate:       data.IssueDate.Format("2006-01-02"),
		DueDate:         data.DueDate.Format("2006-01-02"),
		InvoiceTypeCode: 380,
		DocumentCurrencyCode: data.Currency,
	}

	// Tax category
	taxCategoryID := "S"
	taxExemptionReason := ""
	if !data.SellerVATRegistered {
		taxCategoryID = "E"
		taxExemptionReason = "TAXE - Loss of right - Loss of the right of exemption"
	}

	// Supplier party — tax ID always prefixed with country code
	sellerTaxID := data.SellerCUI
	if sellerTaxID == "" {
		sellerTaxID = data.SellerCNP
	}
	sellerCompanyID := sellerTaxID
	sellerTaxID = prefixCountryCode(sellerTaxID, data.SellerCountryCode)

	inv.AccountingSupplierParty = ublPartyWrapper{
		Party: ublParty{
			PostalAddress: ublAddress{
				StreetName:       data.SellerStreet,
				CityName:         data.SellerCity,
				PostalZone:       data.SellerPostalCode,
				CountrySubentity: formatCountyISO(data.SellerCounty, data.SellerCountryCode),
				Country:          ublCountry{IdentificationCode: data.SellerCountryCode},
			},
			PartyTaxScheme: ublPartyTaxScheme{
				CompanyID: sellerTaxID,
				TaxScheme: ublTaxSchemeID{ID: "VAT"},
			},
			PartyLegalEntity: ublPartyLegalEntity{
				RegistrationName: data.SellerName,
				CompanyID:        sellerCompanyID,
			},
		},
	}

	// Buyer party — tax ID always prefixed with country code
	buyerTaxID := prefixCountryCode(data.BuyerCUI, data.BuyerCountryCode)

	inv.AccountingCustomerParty = ublPartyWrapper{
		Party: ublParty{
			PostalAddress: ublAddress{
				StreetName:       data.BuyerStreet,
				CityName:         data.BuyerCity,
				PostalZone:       data.BuyerPostalCode,
				CountrySubentity: formatCountyISO(data.BuyerCounty, data.BuyerCountryCode),
				Country:          ublCountry{IdentificationCode: data.BuyerCountryCode},
			},
			PartyTaxScheme: ublPartyTaxScheme{
				CompanyID: buyerTaxID,
				TaxScheme: ublTaxSchemeID{ID: "VAT"},
			},
			PartyLegalEntity: ublPartyLegalEntity{
				RegistrationName: data.BuyerName,
				CompanyID:        data.BuyerCUI,
			},
		},
	}

	// Tax total
	inv.TaxTotal = ublTaxTotal{
		TaxAmount: ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.TaxAmount)},
		TaxSubtotal: ublTaxSubtotal{
			TaxableAmount: ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.NetAmount)},
			TaxAmount:     ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.TaxAmount)},
			TaxCategory: ublTaxCategory{
				ID:                   taxCategoryID,
				Percent:              fmt.Sprintf("%.2f", data.TaxPercent),
				TaxExemptionReason:   taxExemptionReason,
				TaxScheme:            ublTaxSchemeID{ID: "VAT"},
			},
		},
	}

	// Monetary total
	inv.LegalMonetaryTotal = ublLegalMonetaryTotal{
		LineExtensionAmount:  ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.NetAmount)},
		TaxExclusiveAmount:   ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.NetAmount)},
		TaxInclusiveAmount:   ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.GrossAmount)},
		PayableAmount:        ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.GrossAmount)},
	}

	// Invoice line
	pricePerNight := data.NetAmount
	if data.Nights > 0 {
		pricePerNight = data.NetAmount / float64(data.Nights)
	}

	inv.InvoiceLine = ublInvoiceLine{
		ID:                  "1",
		InvoicedQuantity:    ublQuantity{UnitCode: "DAY", Value: fmt.Sprintf("%d", data.Nights)},
		LineExtensionAmount: ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", data.NetAmount)},
		Item: ublItem{
			Name: data.Description,
			ClassifiedTaxCategory: ublClassifiedTaxCategory{
				ID:        taxCategoryID,
				Percent:   fmt.Sprintf("%.2f", data.TaxPercent),
				TaxScheme: ublTaxSchemeID{ID: "VAT"},
			},
		},
		Price: ublPrice{
			PriceAmount: ublAmount{CurrencyID: data.Currency, Value: fmt.Sprintf("%.2f", pricePerNight)},
		},
	}

	output, err := xml.MarshalIndent(inv, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}

// UBL 2.1 XML structs

type ublInvoice struct {
	XMLName  xml.Name `xml:"Invoice"`
	XMLNs    string   `xml:"xmlns,attr"`
	XMLNsCac string   `xml:"xmlns:cac,attr"`
	XMLNsCbc string   `xml:"xmlns:cbc,attr"`

	CustomizationID      string `xml:"cbc:CustomizationID"`
	ID                   string `xml:"cbc:ID"`
	IssueDate            string `xml:"cbc:IssueDate"`
	DueDate              string `xml:"cbc:DueDate"`
	InvoiceTypeCode      int    `xml:"cbc:InvoiceTypeCode"`
	DocumentCurrencyCode string `xml:"cbc:DocumentCurrencyCode"`

	AccountingSupplierParty ublPartyWrapper       `xml:"cac:AccountingSupplierParty"`
	AccountingCustomerParty ublPartyWrapper       `xml:"cac:AccountingCustomerParty"`
	TaxTotal                ublTaxTotal           `xml:"cac:TaxTotal"`
	LegalMonetaryTotal      ublLegalMonetaryTotal `xml:"cac:LegalMonetaryTotal"`
	InvoiceLine             ublInvoiceLine        `xml:"cac:InvoiceLine"`
}

type ublPartyWrapper struct {
	Party ublParty `xml:"cac:Party"`
}

type ublParty struct {
	PostalAddress    ublAddress          `xml:"cac:PostalAddress"`
	PartyTaxScheme   ublPartyTaxScheme   `xml:"cac:PartyTaxScheme"`
	PartyLegalEntity ublPartyLegalEntity `xml:"cac:PartyLegalEntity"`
}

type ublAddress struct {
	StreetName       string     `xml:"cbc:StreetName,omitempty"`
	CityName         string     `xml:"cbc:CityName,omitempty"`
	PostalZone       string     `xml:"cbc:PostalZone,omitempty"`
	CountrySubentity string     `xml:"cbc:CountrySubentity,omitempty"`
	Country          ublCountry `xml:"cac:Country"`
}

type ublCountry struct {
	IdentificationCode string `xml:"cbc:IdentificationCode"`
}

type ublPartyTaxScheme struct {
	CompanyID string         `xml:"cbc:CompanyID"`
	TaxScheme ublTaxSchemeID `xml:"cac:TaxScheme"`
}

type ublPartyLegalEntity struct {
	RegistrationName string `xml:"cbc:RegistrationName"`
	CompanyID        string `xml:"cbc:CompanyID,omitempty"`
}

type ublTaxSchemeID struct {
	ID string `xml:"cbc:ID"`
}

type ublTaxTotal struct {
	TaxAmount   ublAmount      `xml:"cbc:TaxAmount"`
	TaxSubtotal ublTaxSubtotal `xml:"cac:TaxSubtotal"`
}

type ublTaxSubtotal struct {
	TaxableAmount ublAmount      `xml:"cbc:TaxableAmount"`
	TaxAmount     ublAmount      `xml:"cbc:TaxAmount"`
	TaxCategory   ublTaxCategory `xml:"cac:TaxCategory"`
}

type ublTaxCategory struct {
	ID                   string         `xml:"cbc:ID"`
	Percent              string         `xml:"cbc:Percent"`
	TaxExemptionReason   string         `xml:"cbc:TaxExemptionReason,omitempty"`
	TaxScheme            ublTaxSchemeID  `xml:"cac:TaxScheme"`
}

type ublLegalMonetaryTotal struct {
	LineExtensionAmount ublAmount `xml:"cbc:LineExtensionAmount"`
	TaxExclusiveAmount  ublAmount `xml:"cbc:TaxExclusiveAmount"`
	TaxInclusiveAmount  ublAmount `xml:"cbc:TaxInclusiveAmount"`
	PayableAmount       ublAmount `xml:"cbc:PayableAmount"`
}

type ublAmount struct {
	CurrencyID string `xml:"currencyID,attr"`
	Value      string `xml:",chardata"`
}

type ublQuantity struct {
	UnitCode string `xml:"unitCode,attr"`
	Value    string `xml:",chardata"`
}

type ublInvoiceLine struct {
	ID                  string                 `xml:"cbc:ID"`
	InvoicedQuantity    ublQuantity            `xml:"cbc:InvoicedQuantity"`
	LineExtensionAmount ublAmount              `xml:"cbc:LineExtensionAmount"`
	Item                ublItem                `xml:"cac:Item"`
	Price               ublPrice               `xml:"cac:Price"`
}

type ublItem struct {
	Name                  string                    `xml:"cbc:Name"`
	ClassifiedTaxCategory ublClassifiedTaxCategory  `xml:"cac:ClassifiedTaxCategory"`
}

type ublClassifiedTaxCategory struct {
	ID        string         `xml:"cbc:ID"`
	Percent   string         `xml:"cbc:Percent"`
	TaxScheme ublTaxSchemeID `xml:"cac:TaxScheme"`
}

type ublPrice struct {
	PriceAmount ublAmount `xml:"cbc:PriceAmount"`
}

// prefixCountryCode ensures a tax identifier has the country code prefix (e.g. "RO12345678").
func prefixCountryCode(id, countryCode string) string {
	if id == "" || countryCode == "" {
		return id
	}
	upper := strings.ToUpper(id)
	if strings.HasPrefix(upper, strings.ToUpper(countryCode)) {
		return upper
	}
	return strings.ToUpper(countryCode) + id
}

// formatCountyISO formats a county code to ISO 3166-2 format (e.g. "RO-B", "RO-BH").
// Accepts full names ("BUCURESTI"), short codes ("B", "BH"), or prefixed ("RO-B").
func formatCountyISO(county, countryCode string) string {
	if county == "" || countryCode == "" {
		return county
	}
	county = strings.TrimSpace(county)
	prefix := strings.ToUpper(countryCode) + "-"

	// Strip existing prefix if present
	upper := strings.ToUpper(county)
	if strings.HasPrefix(upper, prefix) {
		upper = upper[len(prefix):]
	}

	// If it's a full county name, map it to the ISO code
	if code, ok := countyNameToISO[upper]; ok {
		return prefix + code
	}

	return prefix + upper
}
