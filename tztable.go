package main

import (
	"fmt"
)

type TZLabel struct {
	Abbreviation string
	Full         string
	UTCOffset    int
}

// lookupFullTZ returns the full name of an abbreviation of a timezone.
// If two tzs exist for an abbreviation, the closest tz to offset is returned.
func lookupFullTZ(abbrev string, utcOffset int) (TZLabel, error) {
	val, exists := tzAbbrevsTable[abbrev]
	if !exists {
		return TZLabel{}, fmt.Errorf("abbrev: %s does not exist in the lookup table", abbrev)
	}
	ret := val[0]
	if len(val) > 1 { // two possible tzs for this abbrev
		lesserDiff := val[0]
		for i := 1; i < len(val); i++ {
			if abs(utcOffset-val[i].UTCOffset) < abs(utcOffset-lesserDiff.UTCOffset) {
				lesserDiff = val[i]
			}
		}
		ret = lesserDiff
	}
	return ret, nil
}

func abs(x int) int {
	if x < 0 {
		return x * -1
	}
	return x
}

var tzAbbrevsTable = map[string][]TZLabel{
	"ACDT":  {{"ACDT", "Australian Central Daylight Saving Time", 10}},
	"ACST":  {{"ACST", "Australian Central Standard Time", 9}},
	"ACT":   {{"ACT", "Acre Time", -5}},
	"ACWST": {{"ACWST", "Australian Central Western Standard Time (unofficial)", 8}},
	"ADT":   {{"ADT", "Atlantic Daylight Time", -3}},
	"AEDT":  {{"AEDT", "Australian Eastern Daylight Saving Time", 11}},
	"AEST":  {{"AEST", "Australian Eastern Standard Time", 10}},
	"AET":   {{"AET", "Australian Eastern Time", 10}},
	"AFT":   {{"AFT", "Afghanistan Time", 4}},
	"AKDT":  {{"AKDT", "Alaska Daylight Time", -8}},
	"AKST":  {{"AKST", "Alaska Standard Time", -9}},
	"ALMT":  {{"ALMT", "Alma-Ata Time", 6}},
	"AMST":  {{"AMST", "Amazon Summer Time (Brazil)", -3}},
	"AMT":   {{"AMT", "Amazon Time (Brazil)", -4}, {"AMT", "Armenia Time", 4}},
	"ANAT":  {{"ANAT", "Anadyr Time", 12}},
	"AQTT":  {{"AQTT", "Aqtobe Time", 5}},
	"ART":   {{"ART", "Argentina Time", -3}},
	"AST":   {{"AST", "Arabia Standard Time", 3}, {"AST", "Atlantic Standard Time", -4}},
	"AWST":  {{"AWST", "Australian Western Standard Time", 8}},
	"AZOST": {{"AZOST", "Azores Summer Time", 0}},
	"AZOT":  {{"AZOT", "Azores Standard Time", -1}},
	"AZT":   {{"AZT", "Azerbaijan Time", 4}},
	"BNT":   {{"BNT", "Brunei Time", 8}},
	"BIOT":  {{"BIOT", "British Indian Ocean Time", 6}},
	"BIT":   {{"BIT", "Baker Island Time", -12}},
	"BOT":   {{"BOT", "Bolivia Time", -4}},
	"BRST":  {{"BRST", "Brasília Summer Time", -2}},
	"BRT":   {{"BRT", "Brasília Time", -3}},
	"BST":   {{"BST", "Bangladesh Standard Time", 6}, {"BST", "Bougainville Standard Time", 11}},
	"BTT":   {{"BTT", "Bhutan Time", 6}},
	"CAT":   {{"CAT", "Central Africa Time", 2}},
	"CCT":   {{"CCT", "Cocos Islands Time", 6}},
	"CDT":   {{"CDT", "Central Daylight Time (North America)", -5}, {"CDT", "Cuba Daylight Time", -4}},
	"CEST":  {{"CEST", "Central European Summer Time (Cf. HAEC)", 2}},
	"CET":   {{"CET", "Central European Time", 1}},
	"CHADT": {{"CHADT", "Chatham Daylight Time", 13}},
	"CHAST": {{"CHAST", "Chatham Standard Time", 12}},
	"CHOT":  {{"CHOT", "Choibalsan Standard Time", 8}},
	"CHOST": {{"CHOST", "Choibalsan Summer Time", 9}},
	"CHST":  {{"CHST", "Chamorro Standard Time", 10}},
	"CHUT":  {{"CHUT", "Chuuk Time", 10}},
	"CIST":  {{"CIST", "Clipperton Island Standard Time", -8}},
	"CKT":   {{"CKT", "Cook Island Time", -10}},
	"CLST":  {{"CLST", "Chile Summer Time", -3}},
	"CLT":   {{"CLT", "Chile Standard Time", -4}},
	"COST":  {{"COST", "Colombia Summer Time", -4}},
	"COT":   {{"COT", "Colombia Time", -5}},
	"CST":   {{"CST", "Central Standard Time (North America)", -6}, {"CST", "China Standard Time", 8}, {"CST", "Cuba Standard Time", -5}},
	"CT":    {{"CT", "Central Time", -6}},
	"CVT":   {{"CVT", "Cape Verde Time", -1}},
	"CWST":  {{"CWST", "Central Western Standard Time (Australia) unofficial", 8}},
	"CXT":   {{"CXT", "Christmas Island Time", 7}},
	"DAVT":  {{"DAVT", "Davis Time", 7}},
	"DDUT":  {{"DDUT", "Dumont d'Urville Time", 10}},
	"DFT":   {{"DFT", "AIX-specific equivalent of Central European Time", 1}},
	"EASST": {{"EASST", "Easter Island Summer Time", -5}},
	"EAST":  {{"EAST", "Easter Island Standard Time", -6}},
	"EAT":   {{"EAT", "East Africa Time", 3}},
	"ECT":   {{"ECT", "Eastern Caribbean Time", -4}, {"ECT", "Ecuador Time", -5}},
	"EDT":   {{"EDT", "Eastern Daylight Time (North America)", -4}},
	"EEST":  {{"EEST", "Eastern European Summer Time", 3}},
	"EET":   {{"EET", "Eastern European Time", 2}},
	"EGST":  {{"EGST", "Eastern Greenland Summer Time", 0}},
	"EGT":   {{"EGT", "Eastern Greenland Time", -1}},
	"EST":   {{"EST", "Eastern Standard Time (North America)", -5}},
	"FET":   {{"FET", "Further-eastern European Time", 3}},
	"FJT":   {{"FJT", "Fiji Time", 12}},
	"FKST":  {{"FKST", "Falkland Islands Summer Time", -3}},
	"FKT":   {{"FKT", "Falkland Islands Time", -4}},
	"FNT":   {{"FNT", "Fernando de Noronha Time", -2}},
	"GALT":  {{"GALT", "Galápagos Time", -6}},
	"GAMT":  {{"GAMT", "Gambier Islands Time", -9}},
	"GET":   {{"GET", "Georgia Standard Time", 4}},
	"GFT":   {{"GFT", "French Guiana Time", -3}},
	"GILT":  {{"GILT", "Gilbert Island Time", 12}},
	"GIT":   {{"GIT", "Gambier Island Time", -9}},
	"GMT":   {{"GMT", "Greenwich Mean Time", 0}},
	"GST":   {{"GST", "South Georgia and the South Sandwich Islands Time", -2}, {"GST", "Gulf Standard Time", 4}},
	"GYT":   {{"GYT", "Guyana Time", -4}},
	"HDT":   {{"HDT", "Hawaii–Aleutian Daylight Time", -9}},
	"HAEC":  {{"HAEC", "Heure Avancée d'Europe Centrale", 2}},
	"HST":   {{"HST", "Hawaii–Aleutian Standard Time", -10}},
	"HKT":   {{"HKT", "Hong Kong Time", 8}},
	"HMT":   {{"HMT", "Heard and McDonald Islands Time", 5}},
	"HOVST": {{"HOVST", "Hovd Summer Time", 8}},
	"HOVT":  {{"HOVT", "Hovd Time", 7}},
	"ICT":   {{"ICT", "Indochina Time", 7}},
	"IDLW":  {{"IDLW", "International Day Line West", -12}},
	"IDT":   {{"IDT", "Israel Daylight Time", 3}},
	"IOT":   {{"IOT", "Indian Ocean Time", 3}},
	"IRDT":  {{"IRDT", "Iran Daylight Time", 4}},
	"IRKT":  {{"IRKT", "Irkutsk Time", 8}},
	"IRST":  {{"IRST", "Iran Standard Time", 3}},
	"IST":   {{"IST", "Indian Standard Time", 5}, {"IST", "Irish Standard Time", 1}, {"IST", "Israel Standard Time", 2}},
	"JST":   {{"JST", "Japan Standard Time", 9}},
	"KALT":  {{"KALT", "Kaliningrad Time", 2}},
	"KGT":   {{"KGT", "Kyrgyzstan Time", 6}},
	"KOST":  {{"KOST", "Kosrae Time", 11}},
	"KRAT":  {{"KRAT", "Krasnoyarsk Time", 7}},
	"KST":   {{"KST", "Korea Standard Time", 9}},
	"LHST":  {{"LHST", "Lord Howe Standard Time", 10}},
	"LINT":  {{"LINT", "Line Islands Time", 14}},
	"MAGT":  {{"MAGT", "Magadan Time", 12}},
	"MART":  {{"MART", "Marquesas Islands Time", -9}},
	"MAWT":  {{"MAWT", "Mawson Station Time", 5}},
	"MDT":   {{"MDT", "Mountain Daylight Time (North America)", -6}},
	"MET":   {{"MET", "Middle European Time (same zone as CET)", 1}},
	"MEST":  {{"MEST", "Middle European Summer Time (same zone as CEST)", 2}},
	"MHT":   {{"MHT", "Marshall Islands Time", 12}},
	"MIST":  {{"MIST", "Macquarie Island Station Time", 11}},
	"MIT":   {{"MIT", "Marquesas Islands Time", -9}},
	"MMT":   {{"MMT", "Myanmar Standard Time", 6}},
	"MSK":   {{"MSK", "Moscow Time", 3}},
	"MST":   {{"MST", "Malaysia Standard Time", 8}, {"MST", "Mountain Standard Time (North America)", -7}},
	"MUT":   {{"MUT", "Mauritius Time", 4}},
	"MVT":   {{"MVT", "Maldives Time", 5}},
	"MYT":   {{"MYT", "Malaysia Time", 8}},
	"NCT":   {{"NCT", "New Caledonia Time", 11}},
	"NDT":   {{"NDT", "Newfoundland Daylight Time", -2}},
	"NFT":   {{"NFT", "Norfolk Island Time", 11}},
	"NOVT":  {{"NOVT", "Novosibirsk Time", 7}},
	"NPT":   {{"NPT", "Nepal Time", 5}},
	"NST":   {{"NST", "Newfoundland Standard Time", -3}},
	"NT":    {{"NT", "Newfoundland Time", -3}},
	"NUT":   {{"NUT", "Niue Time", -11}},
	"NZDT":  {{"NZDT", "New Zealand Daylight Time", 13}},
	"NZST":  {{"NZST", "New Zealand Standard Time", 12}},
	"OMST":  {{"OMST", "Omsk Time", 6}},
	"ORAT":  {{"ORAT", "Oral Time", 5}},
	"PDT":   {{"PDT", "Pacific Daylight Time (North America)", -7}},
	"PET":   {{"PET", "Peru Time", -5}},
	"PETT":  {{"PETT", "Kamchatka Time", 12}},
	"PGT":   {{"PGT", "Papua New Guinea Time", 10}},
	"PHOT":  {{"PHOT", "Phoenix Island Time", 13}},
	"PHT":   {{"PHT", "Philippine Time", 8}},
	"PKT":   {{"PKT", "Pakistan Standard Time", 5}},
	"PMDT":  {{"PMDT", "Saint Pierre and Miquelon Daylight Time", -2}},
	"PMST":  {{"PMST", "Saint Pierre and Miquelon Standard Time", -3}},
	"PONT":  {{"PONT", "Pohnpei Standard Time", 11}},
	"PST":   {{"PST", "Pacific Standard Time (North America)", -8}, {"PST", "Philippine Standard Time", 8}},
	"PWT":   {{"PWT", "Palau Time", 9}},
	"PYST":  {{"PYST", "Paraguay Summer Time", -3}},
	"PYT":   {{"PYT", "Paraguay Time", -4}},
	"RET":   {{"RET", "Réunion Time", 4}},
	"ROTT":  {{"ROTT", "Rothera Research Station Time", -3}},
	"SAKT":  {{"SAKT", "Sakhalin Island Time", 11}},
	"SAMT":  {{"SAMT", "Samara Time", 4}},
	"SAST":  {{"SAST", "South African Standard Time", 2}},
	"SBT":   {{"SBT", "Solomon Islands Time", 11}},
	"SCT":   {{"SCT", "Seychelles Time", 4}},
	"SDT":   {{"SDT", "Samoa Daylight Time", -10}},
	"SGT":   {{"SGT", "Singapore Time", 8}},
	"SLST":  {{"SLST", "Sri Lanka Standard Time", 5}},
	"SRET":  {{"SRET", "Srednekolymsk Time", 11}},
	"SRT":   {{"SRT", "Suriname Time", -3}},
	"SST":   {{"SST", "Samoa Standard Time", -11}, {"SST", "Singapore Standard Time", 8}},
	"SYOT":  {{"SYOT", "Showa Station Time", 3}},
	"TAHT":  {{"TAHT", "Tahiti Time", -10}},
	"THA":   {{"THA", "Thailand Standard Time", 7}},
	"TFT":   {{"TFT", "French Southern and Antarctic Time", 5}},
	"TJT":   {{"TJT", "Tajikistan Time", 5}},
	"TKT":   {{"TKT", "Tokelau Time", 13}},
	"TLT":   {{"TLT", "Timor Leste Time", 9}},
	"TMT":   {{"TMT", "Turkmenistan Time", 5}},
	"TRT":   {{"TRT", "Turkey Time", 3}},
	"TOT":   {{"TOT", "Tonga Time", 13}},
	"TVT":   {{"TVT", "Tuvalu Time", 12}},
	"ULAST": {{"ULAST", "Ulaanbaatar Summer Time", 9}},
	"ULAT":  {{"ULAT", "Ulaanbaatar Standard Time", 8}},
	"UTC":   {{"UTC", "Coordinated Universal Time", 0}},
	"UYST":  {{"UYST", "Uruguay Summer Time", -2}},
	"UYT":   {{"UYT", "Uruguay Standard Time", -3}},
	"UZT":   {{"UZT", "Uzbekistan Time", 5}},
	"VET":   {{"VET", "Venezuelan Standard Time", -4}},
	"VLAT":  {{"VLAT", "Vladivostok Time", 10}},
	"VOLT":  {{"VOLT", "Volgograd Time", 4}},
	"VOST":  {{"VOST", "Vostok Station Time", 6}},
	"VUT":   {{"VUT", "Vanuatu Time", 11}},
	"WAKT":  {{"WAKT", "Wake Island Time", 12}},
	"WAST":  {{"WAST", "West Africa Summer Time", 2}},
	"WAT":   {{"WAT", "West Africa Time", 1}},
	"WEST":  {{"WEST", "Western European Summer Time", 1}},
	"WET":   {{"WET", "Western European Time", 0}},
	"WIB":   {{"WIB", "Western Indonesian Time", 7}},
	"WIT":   {{"WIT", "Eastern Indonesian Time", 9}},
	"WITA":  {{"WITA", "Central Indonesia Time", 8}},
	"WGST":  {{"WGST", "West Greenland Summer Time", -2}},
	"WGT":   {{"WGT", "West Greenland Time", -3}},
	"WST":   {{"WST", "Western Standard Time", 8}},
	"YAKT":  {{"YAKT", "Yakutsk Time", 9}},
	"YEKT":  {{"YEKT", "Yekaterinburg Time", 5}},
}

// // https://en.wikipedia.org/wiki/List_of_time_zone_abbreviations
// let t = ""
// Array.from(document.getElementsByTagName("tbody")[1].children).forEach((el) => {
//     let ar = el.children
//     let abbrev = ar[0].textContent.replaceAll("\n","")
//     let full = ar[1].textContent.replaceAll("\n","")
//         .replaceAll(/\[.*\]/g, "") // remove citations e.g, [5]
//     let of = ar[2].textContent
//         .replaceAll(/[\/–:].*/g,"") // replace everything after "/" or emdash "–" or ":" aka unneccesary tz info to calculate most accurate tz
//         .replaceAll("−", "-")  // replace minus sign with a parsable minus sign
//         .replaceAll("±", "")
//         .replaceAll("UTC", "")
//         .replaceAll("\n","")
//     of = Number.parseInt(of)
//     t += `"${abbrev}": {{"${abbrev}", "${full}", ${of}}},` + "\n"
// })
// console.log(t)
