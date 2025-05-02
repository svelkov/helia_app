package validation

// PartneriValidator implements validation for Partneri entities.
type PartneriValidator struct{}

// NewPartneriValidator creates a new instance of PartneriValidator.
func NewPartneriValidator() *PartneriValidator {
	return &PartneriValidator{}
}

func PartneriValidationRules() []ValidationRule {
	return []ValidationRule{
		{
			Field:   "sifra",
			Message: "Morate uneti analitiku sifru partnera!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "naziv",
			Message: "Morate uneti naziv partnera!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "mesto",
			Message: "Morate uneti naziv mesta!!!",
			Check: func(value any) bool {
				str, ok := value.(string)
				return ok && len(str) > 0
			},
		},
		{
			Field:   "tippdv",
			Message: "Morate izabrati u koji sistem PDV-a pripada tip partnera!!!",
			Check: func(value any) bool {
				val, ok := value.(int)
				return ok && !(val >= 1 && val <= 5)
			},
		},
	}
}

// sPomKONTO is string = ""

// IF gsWindowMode = "Create" THEN
// 	HReadSeekFirst(PARTNERI,GODKARSIFRA,[gnGOD,gnKAR,EDT_SIFRA],hLockNo)
// 	IF HFound(PARTNERI) THEN
// 		lOK = False
// 		Info("Partner sa sifrom:" + EDT_SIFRA + " vec postoji. Nedozvoljen dupli unos!!!")
// 		ReturnToCapture(EDT_SIFRA)
// 	END
// END

// //IF EDT_JIB = "" THEN
// //	lOK = False
// //	Info("Morate uneti JIB!!!")
// //	ReturnToCapture(EDT_JIB)
// //END
// //IF EDT_MATBR = "" THEN
// //	lOK = False
// //	Info("Morate uneti maticni broj partnera!!!")
// //	ReturnToCapture(EDT_MATBR)
// //END
// //if edt_jmbg <> "" then
// //	nJMBG = EDT_JMBG
// //	if nJMBG modulo 13 <> 0 then
// //		lOK = False
// //		Info("Neispravan JMBG!!!")
// //		edt_jmbg = ""
// //		ReturnToCapture(edt_JMBG)
// //	END
// //END

// //IF EDT_ZIRO = "" THEN
// //	lOK = False
// //	Info("Morate uneti broj tekuceg racuna!!!")
// //	ReturnToCapture(EDT_ZIRO)
// //END
// IF COMBO_TIP = "" THEN
// 	lOK = False
// 	Info("Morate izabrati u koji sistem PDV-a pripada tip partnera!!!")
// 	ReturnToCapture(COMBO_TIP)
// END

// IF (COMBO_TIP = 1 OR COMBO_TIP = 2) AND EDT_PIB <> "" AND EDT_PIB <> nOLDPIB THEN
// 	HReadSeekFirst(PARTNERI,GODKARPIB,[gnGOD,gnKAR,EDT_PIB],hLockNo)
// 	IF HFound(PARTNERI) THEN
// 		//1 : Da
// 		//2 : Ne
// 		SWITCH Dialog("Partner sa izabranim PIB-om već postoji!. ")
// 			// Da
// 			CASE 1
// 				lOK = True
// 			// Ne
// 			CASE 2
// 				lOK = False
// 				Info("Partner sa ovakvim PIB-om vec postoji!!!")
// 				ReturnToCapture(EDT_PIB)
// 		END
// 	END
// END

// IF COMBO_TIP = 3 AND EDT_BPG <> "" AND EDT_BPG <> nOLDBPG THEN
// 	HReadSeekFirst(PARTNERI,GODKARBPG,[gnGOD,gnKAR,EDT_BPG],hLockNo)
// 	IF HFound(PARTNERI) THEN
// 		//1 : Da
// 		//2 : Ne
// 		SWITCH Dialog("Partner sa ovakvim BPG-om već postoji! ")
// 			// Da
// 			CASE 1
// 				lOK = True
// 			// Ne
// 			CASE 2
// 				lOK = False
// 				Info("Partner sa ovakvim BPG-om vec postoji!!!")
// 				ReturnToCapture(EDT_BPG)
// 		END
// 	END
// END

// IF COMBO_TIP = 4 AND EDT_JMBG <> "" AND EDT_JMBG <> nOLDJMBG THEN
// 	HReadSeekFirst(PARTNERI,GODKARJMBG,[gnGOD,gnKAR,EDT_JMBG],hLockNo)
// 	IF HFound(PARTNERI) THEN
// 		//1 : Da
// 		//2 : Ne
// 		SWITCH Dialog("Partner sa ovakvim JMBG-om već postoji!")
// 			// Da
// 			CASE 1
// 				lOK = True
// 			// Ne
// 			CASE 2
// 				lOK = False
// 				Info("Partner sa ovakvim JMBG-om veæ postoji!!!")
// 				ReturnToCapture(EDT_JMBG)
// 		END
// 	END
// END

// IF COMBO_TIP = 5 THEN
// 	IF EDT_JMBG <> "" AND EDT_JMBG <> nOLDJMBG THEN
// 		HReadSeekFirst(PARTNERI,GODKARJMBG,[gnGOD,gnKAR,EDT_JMBG],hLockNo)
// 		IF HFound(PARTNERI) THEN
// 			//1 : Da
// 			//2 : Ne
// 			//1 :
// 			SWITCH Dialog("Student sa ovakvim JMBG-om već postoji!")
// 				// Da
// 				CASE 1
// 					lOK1 = True
// 				// Ne
// 				CASE 2
// 					lOK1 = False
// 					Info("Student sa ovakvim JMBG-om veæ postoji!!!")
// 					ReturnToCapture(EDT_JMBG)
// 			END
// 		END
// 	END
// 	IF EDT_INDEX <> "" AND EDT_INDEX <> sOLDINDEX THEN
// 		HReadSeekFirst(PARTNERI,GODKARINDEX,[gnGOD,gnKAR,EDT_INDEX],hLockNo)
// 		IF HFound(PARTNERI) THEN
// 			//1 : Da
// 			//2 : Ne
// 			SWITCH Dialog("Student sa ovakvim brojem indeksa postoji!")
// 				// Da
// 				CASE 1
// 					lOK2 = True
// 				// Ne
// 				CASE 2
// 					lOK2 = False
// 					Info("Student sa ovakvim brojem indeksa vec postoji!!!")
// 					ReturnToCapture(EDT_INDEX)
// 			END
// 		END
// 	END
// 	IF lOK1 = True AND lOK2 = True THEN
// 		lOK = True
// 	ELSE
// 		lOK = False
// 		Info("Student sa ovakvim podacima veæ postoji!!!")
// 		ReturnToCapture(EDT_JMBG)
// 	END
// END

// // PROVERA PIB-a
// //*********************************************************
// //ISO 7064, MODUL (11,10)
// //1. prva znamenka zbroji se s brojem 10
// //2. dobiveni zbroj cjelobrojno (s ostatkom) podijeli se brojem 10; ako je dobiveni
// //ostatak 0 zamijeni se brojem 10 (ovaj broj je tzv. meduostatak)
// //3. dobiveni meduostatak pomnoži se brojem 2
// //4. dobiveni umnožak cjelobrojno (s ostatkom) podijeli se brojem 11; ovaj ostatak
// //matematicki nikako ne može biti 0 jer je rezultat prethodnog koraka uvijek paran
// //broj
// //5. slijedeca znamenka zbroji se s ostatkom u prethodnom koraku
// //6. ponavljaju se koraci 2, 3, 4 i 5 dok se ne potroše sve znamenke
// //7. razlika izmedu broja 11 i ostatka u zadnjem koraku je kontrolna znamenka

// nCifra is int
// i is int
// nZbir is int
// nOstatak is int=10
// nMedjuostatak is int
// nKontrol is int

// IF RADIO_Radio1 <= 2 AND (COMBO_TIP = 1 OR COMBO_TIP = 2) THEN
// 	IF (EDT_PIB="" OR (Length(EDT_PIB) <> 9)) AND EDT_JMBG = "" THEN
// 		lOK = False
// 		Info("Vas PIB nema 9 cifara ili ima više od 9 cifara, upišite novi")
// 	END
// 	IF Length(EDT_PIB) = 9 THEN
// 		FOR i= 1 _TO_ Length(EDT_PIB)-1
// 			nCifra=Middle(EDT_PIB,i,1)
// 			nZbir=nCifra+nOstatak
// 			nMedjuostatak=modulo(nZbir,10)
// 			IF nMedjuostatak=0 THEN
// 				nMedjuostatak=10
// 			END
// 			nOstatak=modulo(nMedjuostatak*2,11)
// 		END

// 		// dodao stole 05.03.2011. nije radilo korektno
// 		nKontrol= modulo((11 - nOstatak),10)

// 		IF Right(EDT_PIB,1)<>nKontrol THEN
// 			lOK=False
// 			Info("Uneti PIB nije korektan!!!","Unesite korektan PIB...")
// 		END
// 	END
// 	IF EDT_PIB="" AND EDT_JMBG <> "" AND EDT_JMBG <> nOLDJMBG THEN
// 		HReadSeekFirst(PARTNERI,GODKARJMBG,[gnGOD,gnKAR,EDT_JMBG],hLockNo)
// 		IF HFound(PARTNERI) THEN
// 			//1 : Da
// 			//2 : Ne
// 			SWITCH Dialog("Partner sa ovakvim JMBG-om već postoji!")
// 				// Da
// 				CASE 1
// 					lOK = True
// 				// Ne
// 				CASE 2
// 					lOK = False
// 					Info("Partner sa ovakvim JMBG-om veæ postoji!!!")
// 					ReturnToCapture(EDT_JMBG)
// 			END
// 		END
// 	END
// END
// //*********************************************************
// //if gsWindowMode = "Modify" then
// //	IF Position(EDT_KONTA,sOldKONTA,1,WholeWord) <> 1 AND sOldKONTA <> "" THEN
// //		lOK = False
// //		Info("Konta partnera mora da sadrzi vrednost :"+sOldKONTA+" plus nova konta ukoliko se otvara analitika za dodatna klonta...")
// //		EDT_KONTA = sOldKONTA
// //		ReturnToCapture(EDT_KONTA)
// //	END
// //end
// IF gsWindowMode = "Modify" THEN
// 	// proveri da li konta iz prometa postoje u novi set konta
// 	okPromet is boolean = True
// 	ProveraKONTA()
// 	okPromet = True
// 	IF sKontoPromet <> "" THEN
// 		sPomKONTO = ExtractString(sKontoPromet,firstRank,",")  //uzmi prvo konto
// 		// Browse all the sub-strings
// 		WHILE sPomKONTO <> EOT
// 			IF Position(EDT_KONTA,sPomKONTO,1,WholeWord) = 0 AND sPomKONTO <> "" THEN
// 				okPromet = False
// 			END
// 			sPomKONTO = ExtractString(sKontoPromet, nextRank, ",")  //uzmi sledece konto
// 		END	//while
// 	END
// 	IF NOT okPromet THEN
// 		lOK = False
// 		Info("Polje Konta partnera mora da sadrzi sledeca konta :"+sKontoPromet+" zato sto za ova konta postoji promet....")
// 		ReturnToCapture(EDT_KONTA)
// 	END
// END

// IF CBOX_Budzetski = True THEN
// 	IF EDT_JBKJS = "" OR Length(EDT_JBKJS) < 5 THEN
// 		lOK = False
// 		Info("Morate uneti JBKJS od 5 cifara!!!")
// 		ReturnToCapture(EDT_JBKJS)
// 	END
// END
// //IF gsWindowMode = "Modify" THEN
// //	IF soldKonta <> edt_konta and length(sOldkonta) > length(EDT_KONTA) THEN
// //		//1 : Da
// //		//2 : Ne
// //		SWITCH Dialog("")
// //			// Da
// //			CASE 1
// //				// ako je prethodno otvorena analitika za ovog partnera, prilikom izmene ne zeli se ta alitika
// //				// onda obrisi j iz kontnog plana
// //				BrisiANALITIKU(EDT_KONTA)
// //				lok = True
// //			// Ne
// //			CASE 2
// //				lOK = True
// //				EDT_KONTA = sOldKONTA
// //				ReturnToCapture(EDT_KONTA)
// //		END
// //	END
// //END
// //IF EDT_TER = 0 THEN
// //	lOK = False
// //	Info("Morate izabrati teritoriju gde pripada partner!!!")
// //	ReturnToCapture(EDT_TER)
// //END

// //HReadSeekFirst(SIFOP,IDSIFOP,COMBO_IDSIFOP,hLockNo)
// //IF HOut(SIFOP) THEN
// //	lOK = False
// //	Info("Opstina sa takvom sifrom ne posotji!!!")
// //	ReturnToCapture(COMBO_IDSIFOP)
// //END

// //ovde dolazi kod za validaciju. Polja koja su obavezna za unos kontrolisu se ovde.
// //takodje ovde se kontrolise i da li slog koji se unosi vec postoji!!!
