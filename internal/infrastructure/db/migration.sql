-- Database migrations

CREATE OR REPLACE FUNCTION fn_kopiraj_nalog_sa_stavkama(
	p_old_idfnal bigint,
	p_new_nalog bigint,
	p_new_tipdok text,
	p_new_danal date,
	p_new_datob date,
	p_new_opis text,
    p_new_idtipdok integer,
	p_god integer,
	p_kar integer,
	p_user text DEFAULT NULL
)
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
	v_new_idfnal bigint;
	v_new_idfpro bigint;
	v_new_idkir bigint;
	v_new_idkpr bigint;
	r_fpro fpro%ROWTYPE;
BEGIN
	-- Copy FNAL header.
	INSERT INTO fnal (
		god, kar, nalog, tipdok, danal, datob, opis, dug, pot, rbr, brst, abr, nalsts, xdatunosa, xopunos, idtipdok
	)
	SELECT
		p_god,
		p_kar,
		p_new_nalog,
		p_new_tipdok,
		p_new_danal,
		p_new_datob,
		COALESCE(NULLIF(p_new_opis, ''), f.opis),
		f.dug,
		f.pot,
		f.rbr,
		f.brst,
		f.abr,
		f.nalsts,
        NOW(),
        COALESCE(NULLIF(p_user, ''), f.xopunos),
        p_new_idtipdok
	FROM fnal f
	WHERE f.idfnal = p_old_idfnal
	RETURNING idfnal INTO v_new_idfnal;

	IF v_new_idfnal IS NULL THEN
		RAISE EXCEPTION 'Source FNAL not found for idfnal=%', p_old_idfnal;
	END IF;

	-- Copy FPRO rows one-by-one so we can attach new KIR/KPR rows per copied FPRO.
	FOR r_fpro IN
		SELECT *
		FROM fpro
		WHERE idfnal = p_old_idfnal
		ORDER BY rbr, idfpro
	LOOP
		INSERT INTO fpro (
			god, kar, nalog, tipdok, rbr, danal, brst, iznos, kat, opis, dadok,
			rok, vrd, vkonta, konto, sifra, tra, deviznos, kurs, sifval, mi,
			ostatak, dokum, vkrbr, vktip, xdatunosa, xdatizmene, xopunos,
			xopizmene, idfnal, idfvknjrac, idorgjed, idfkpl, mag, komid, rdokid,
			mtroska, mestotrid, flegkomp, dokumv, dadokv, travez, starikonto,
			ojozn, sifkom, kolic, cena, idkir, idkpr
		)
		VALUES (
			p_god,
			p_kar,
			p_new_nalog,
			p_new_tipdok,
			r_fpro.rbr,
			p_new_danal,
			r_fpro.brst,
			r_fpro.iznos,
			r_fpro.kat,
			r_fpro.opis,
			r_fpro.dadok,
			r_fpro.rok,
			r_fpro.vrd,
			r_fpro.vkonta,
			r_fpro.konto,
			r_fpro.sifra,
			r_fpro.tra,
			r_fpro.deviznos,
			r_fpro.kurs,
			r_fpro.sifval,
			r_fpro.mi,
			r_fpro.ostatak,
			r_fpro.dokum,
			r_fpro.vkrbr,
			r_fpro.vktip,
			NOW(),
			NULL,
			COALESCE(NULLIF(p_user, ''), r_fpro.xopunos),
			r_fpro.xopizmene,
			v_new_idfnal,
			r_fpro.idfvknjrac,
			r_fpro.idorgjed,
			r_fpro.idfkpl,
			r_fpro.mag,
			r_fpro.komid,
			r_fpro.rdokid,
			r_fpro.mtroska,
			r_fpro.mestotrid,
			r_fpro.flegkomp,
			r_fpro.dokumv,
			r_fpro.dadokv,
			r_fpro.travez,
			r_fpro.starikonto,
			r_fpro.ojozn,
			r_fpro.sifkom,
			r_fpro.kolic,
			r_fpro.cena,
			NULL,
			NULL
		)
		RETURNING idfpro INTO v_new_idfpro;

		-- VRD=10 -> copy KIR and link copied FPRO.idkir.
		IF r_fpro.vrd = 10 THEN
			v_new_idkir := NULL;

			WITH src AS (
				SELECT *
				FROM kir k
				WHERE
					(r_fpro.idkir IS NOT NULL AND k.idkir = r_fpro.idkir)
					OR (r_fpro.idkir IS NULL AND k.idfpro = r_fpro.idfpro)
				ORDER BY k.idkir
				LIMIT 1
			)
			INSERT INTO kir (
				god, kar, vktip, vkrbr, krbr, idpartneri,
				dknjiz, danal, dizd, kracun,
				iznsapdv, oslobcl24, oslobcl25, izvozsapr, izvozbezpr,
				osn1, pdv1, osn2, pdv2, prom1, prom2,
				nalog, tipdok, vrd, dokum, vpr, rkar, brst,
				konto, sifra, pib, naziv,
				xdatunosa, xdatizmene, xopunos, xopizmene,
				idfvknjrac, idfpro, numdok, datprometa
			)
			SELECT
				p_god,
				p_kar,
				src.vktip,
				src.vkrbr,
				src.krbr,
				src.idpartneri,
				p_new_datob,
				p_new_danal,
				src.dizd,
				src.kracun,
				src.iznsapdv,
				src.oslobcl24,
				src.oslobcl25,
				src.izvozsapr,
				src.izvozbezpr,
				src.osn1,
				src.pdv1,
				src.osn2,
				src.pdv2,
				src.prom1,
				src.prom2,
				p_new_nalog,
				p_new_tipdok,
				src.vrd,
				src.dokum,
				src.vpr,
				src.rkar,
				src.brst,
				src.konto,
				src.sifra,
				src.pib,
				src.naziv,
				NOW(),
				NULL,
				COALESCE(NULLIF(p_user, ''), src.xopunos),
				src.xopizmene,
				src.idfvknjrac,
				v_new_idfpro,
				src.numdok,
				src.datprometa
			FROM src
			RETURNING idkir INTO v_new_idkir;

			IF v_new_idkir IS NOT NULL THEN
				UPDATE fpro
				SET idkir = v_new_idkir
				WHERE idfpro = v_new_idfpro;
			END IF;

		-- VRD=20 -> copy KPR and link copied FPRO.idkpr.
		ELSIF r_fpro.vrd = 20 THEN
			v_new_idkpr := NULL;

			WITH src AS (
				SELECT *
				FROM kpr k
				WHERE
					(r_fpro.idkpr IS NOT NULL AND k.idkpr = r_fpro.idkpr)
					OR (r_fpro.idkpr IS NULL AND k.idfpro = r_fpro.idfpro)
				ORDER BY k.idkpr
				LIMIT 1
			)
			INSERT INTO kpr (
				god, kar, vktip, vkrbr, drbr,
				dknjiz, duvoz, dizd,
				iznsapdv, iznoslob, nisuobvpdv, uvozbezpdv,
				prethpdv, pretpdv1, pretpdv2, uvozpdv,
				poljvred, poljpdv,
				vrd, konto, sifra, ter, uvozosnpdv,
				vpr, osnbezpdv, brst, rkar,
				nalog, dokum, pib, naziv, idpartneri,
				danal, tipdok,
				xdatunosa, xdatizmene, xopunos, xopizmene,
				idfpro, idfvknjrac,
				osnbezpod, datprometa,
				osnovicavt, osnovicant, prethpdvvt, prethpdvnt,
				tkonto, tsifra,
				pretpdv1vt, pretpdv1nt, pretpdv2vt, pretpdv2nt,
				irn, brojirn, demo, individualvatid, vatrecordingstatus,
				datum_stat_indvat, obj, beznak, epp_polje, fseppid
			)
			SELECT
				p_god,
				p_kar,
				src.vktip,
				src.vkrbr,
				src.drbr,
				p_new_datob,
				src.duvoz,
				src.dizd,
				src.iznsapdv,
				src.iznoslob,
				src.nisuobvpdv,
				src.uvozbezpdv,
				src.prethpdv,
				src.pretpdv1,
				src.pretpdv2,
				src.uvozpdv,
				src.poljvred,
				src.poljpdv,
				src.vrd,
				src.konto,
				src.sifra,
				src.ter,
				src.uvozosnpdv,
				src.vpr,
				src.osnbezpdv,
				src.brst,
				src.rkar,
				p_new_nalog,
				src.dokum,
				src.pib,
				src.naziv,
				src.idpartneri,
				p_new_danal,
				p_new_tipdok,
				NOW(),
				NULL,
				COALESCE(NULLIF(p_user, ''), src.xopunos),
				src.xopizmene,
				v_new_idfpro,
				src.idfvknjrac,
				src.osnbezpod,
				src.datprometa,
				src.osnovicavt,
				src.osnovicant,
				src.prethpdvvt,
				src.prethpdvnt,
				src.tkonto,
				src.tsifra,
				src.pretpdv1vt,
				src.pretpdv1nt,
				src.pretpdv2vt,
				src.pretpdv2nt,
				src.irn,
				src.brojirn,
				src.demo,
				src.individualvatid,
				src.vatrecordingstatus,
				src.datum_stat_indvat,
				src.obj,
				src.beznak,
				src.epp_polje,
				src.fseppid
			FROM src
			RETURNING idkpr INTO v_new_idkpr;

			IF v_new_idkpr IS NOT NULL THEN
				UPDATE fpro
				SET idkpr = v_new_idkpr
				WHERE idfpro = v_new_idfpro;
			END IF;
		END IF;
	END LOOP;

	RETURN v_new_idfnal;
END;
$$;

CREATE OR REPLACE FUNCTION fn_storniraj_nalog_sa_stavkama(
	p_old_idfnal bigint,
	p_new_nalog bigint,
	p_new_tipdok text,
	p_new_danal date,
	p_new_datob date,
	p_new_opis text,
	p_god integer,
	p_kar integer,
	p_user text DEFAULT NULL
)
RETURNS bigint
LANGUAGE plpgsql
AS $$
DECLARE
	v_new_idfnal bigint;
	v_new_idfpro bigint;
	v_new_idkir bigint;
	v_new_idkpr bigint;
	r_fpro fpro%ROWTYPE;
BEGIN
	-- Create stornirani FNAL header.
	INSERT INTO fnal (
		god, kar, nalog, tipdok, danal, datob, opis, dug, pot, rbr, brst, abr, nalsts, xdatunosa, xopunos, idtipdok
	)
	SELECT
		p_god,
		p_kar,
		p_new_nalog,
		p_new_tipdok,
		p_new_danal,
		p_new_datob,
		COALESCE(NULLIF(p_new_opis, ''), f.opis),
		-f.dug,
		-f.pot,
		f.rbr,
		f.brst,
		f.abr,
		f.nalsts,
		NOW(),
		COALESCE(NULLIF(p_user, ''), f.xopunos),
		f.idtipdok
	FROM fnal f
	WHERE f.idfnal = p_old_idfnal
	RETURNING idfnal INTO v_new_idfnal;

	IF v_new_idfnal IS NULL THEN
		RAISE EXCEPTION 'Source FNAL not found for idfnal=%', p_old_idfnal;
	END IF;

	-- Copy FPRO rows with storniranje rules and attach KIR/KPR rows.
	FOR r_fpro IN
		SELECT *
		FROM fpro
		WHERE idfnal = p_old_idfnal
		ORDER BY rbr, idfpro
	LOOP
		INSERT INTO fpro (
			god, kar, nalog, tipdok, rbr, danal, brst, iznos, kat, opis, dadok,
			rok, vrd, vkonta, konto, sifra, tra, deviznos, kurs, sifval, mi,
			ostatak, dokum, vkrbr, vktip, xdatunosa, xdatizmene, xopunos,
			xopizmene, idfnal, idfvknjrac, idorgjed, idfkpl, mag, komid, rdokid,
			mtroska, mestotrid, flegkomp, dokumv, dadokv, travez, starikonto,
			ojozn, sifkom, kolic, cena, idkir, idkpr
		)
		VALUES (
			p_god,
			p_kar,
			p_new_nalog,
			p_new_tipdok,
			r_fpro.rbr,
			p_new_danal,
			r_fpro.brst,
			CASE
				WHEN r_fpro.kat IN (1, 3) THEN -r_fpro.iznos
				WHEN r_fpro.kat IN (2, 4) THEN ABS(r_fpro.iznos)
				ELSE r_fpro.iznos
			END,
			CASE
				WHEN r_fpro.kat = 1 THEN 2
				WHEN r_fpro.kat = 2 THEN 1
				WHEN r_fpro.kat = 3 THEN 4
				WHEN r_fpro.kat = 4 THEN 3
				ELSE r_fpro.kat
			END,
			r_fpro.opis,
			r_fpro.dadok,
			r_fpro.rok,
			r_fpro.vrd,
			r_fpro.vkonta,
			r_fpro.konto,
			r_fpro.sifra,
			r_fpro.tra,
			r_fpro.deviznos,
			r_fpro.kurs,
			r_fpro.sifval,
			r_fpro.mi,
			r_fpro.ostatak,
			r_fpro.dokum,
			r_fpro.vkrbr,
			r_fpro.vktip,
			NOW(),
			NULL,
			COALESCE(NULLIF(p_user, ''), r_fpro.xopunos),
			r_fpro.xopizmene,
			v_new_idfnal,
			r_fpro.idfvknjrac,
			r_fpro.idorgjed,
			r_fpro.idfkpl,
			r_fpro.mag,
			r_fpro.komid,
			r_fpro.rdokid,
			r_fpro.mtroska,
			r_fpro.mestotrid,
			r_fpro.flegkomp,
			r_fpro.dokumv,
			r_fpro.dadokv,
			r_fpro.travez,
			r_fpro.starikonto,
			r_fpro.ojozn,
			r_fpro.sifkom,
			r_fpro.kolic,
			r_fpro.cena,
			NULL,
			NULL
		)
		RETURNING idfpro INTO v_new_idfpro;

		-- VRD=10 -> copy KIR and link copied FPRO.idkir.
		IF r_fpro.vrd = 10 THEN
			v_new_idkir := NULL;

			WITH src AS (
				SELECT *
				FROM kir k
				WHERE
					(r_fpro.idkir IS NOT NULL AND k.idkir = r_fpro.idkir)
					OR (r_fpro.idkir IS NULL AND k.idfpro = r_fpro.idfpro)
				ORDER BY k.idkir
				LIMIT 1
			)
			INSERT INTO kir (
				god, kar, vktip, vkrbr, krbr, idpartneri,
				dknjiz, danal, dizd, kracun,
				iznsapdv, oslobcl24, oslobcl25, izvozsapr, izvozbezpr,
				osn1, pdv1, osn2, pdv2, prom1, prom2,
				nalog, tipdok, vrd, dokum, vpr, rkar, brst,
				konto, sifra, pib, naziv,
				xdatunosa, xdatizmene, xopunos, xopizmene,
				idfvknjrac, idfpro, numdok, datprometa
			)
			SELECT
				p_god,
				p_kar,
				src.vktip,
				src.vkrbr,
				src.krbr,
				src.idpartneri,
				p_new_datob,
				p_new_danal,
				src.dizd,
				src.kracun,
				src.iznsapdv,
				src.oslobcl24,
				src.oslobcl25,
				src.izvozsapr,
				src.izvozbezpr,
				src.osn1,
				src.pdv1,
				src.osn2,
				src.pdv2,
				src.prom1,
				src.prom2,
				p_new_nalog,
				p_new_tipdok,
				src.vrd,
				src.dokum,
				src.vpr,
				src.rkar,
				src.brst,
				src.konto,
				src.sifra,
				src.pib,
				src.naziv,
				NOW(),
				NULL,
				COALESCE(NULLIF(p_user, ''), src.xopunos),
				src.xopizmene,
				src.idfvknjrac,
				v_new_idfpro,
				src.numdok,
				src.datprometa
			FROM src
			RETURNING idkir INTO v_new_idkir;

			IF v_new_idkir IS NOT NULL THEN
				UPDATE fpro
				SET idkir = v_new_idkir
				WHERE idfpro = v_new_idfpro;
			END IF;

		-- VRD=20 -> copy KPR and link copied FPRO.idkpr.
		ELSIF r_fpro.vrd = 20 THEN
			v_new_idkpr := NULL;

			WITH src AS (
				SELECT *
				FROM kpr k
				WHERE
					(r_fpro.idkpr IS NOT NULL AND k.idkpr = r_fpro.idkpr)
					OR (r_fpro.idkpr IS NULL AND k.idfpro = r_fpro.idfpro)
				ORDER BY k.idkpr
				LIMIT 1
			)
			INSERT INTO kpr (
				god, kar, vktip, vkrbr, drbr,
				dknjiz, duvoz, dizd,
				iznsapdv, iznoslob, nisuobvpdv, uvozbezpdv,
				prethpdv, pretpdv1, pretpdv2, uvozpdv,
				poljvred, poljpdv,
				vrd, konto, sifra, ter, uvozosnpdv,
				vpr, osnbezpdv, brst, rkar,
				nalog, dokum, pib, naziv, idpartneri,
				danal, tipdok,
				xdatunosa, xdatizmene, xopunos, xopizmene,
				idfpro, idfvknjrac,
				osnbezpod, datprometa,
				osnovicavt, osnovicant, prethpdvvt, prethpdvnt,
				tkonto, tsifra,
				pretpdv1vt, pretpdv1nt, pretpdv2vt, pretpdv2nt,
				irn, brojirn, demo, individualvatid, vatrecordingstatus,
				datum_stat_indvat, obj, beznak, epp_polje, fseppid
			)
			SELECT
				p_god,
				p_kar,
				src.vktip,
				src.vkrbr,
				src.drbr,
				p_new_datob,
				src.duvoz,
				src.dizd,
				src.iznsapdv,
				src.iznoslob,
				src.nisuobvpdv,
				src.uvozbezpdv,
				src.prethpdv,
				src.pretpdv1,
				src.pretpdv2,
				src.uvozpdv,
				src.poljvred,
				src.poljpdv,
				src.vrd,
				src.konto,
				src.sifra,
				src.ter,
				src.uvozosnpdv,
				src.vpr,
				src.osnbezpdv,
				src.brst,
				src.rkar,
				p_new_nalog,
				src.dokum,
				src.pib,
				src.naziv,
				src.idpartneri,
				p_new_danal,
				p_new_tipdok,
				NOW(),
				NULL,
				COALESCE(NULLIF(p_user, ''), src.xopunos),
				src.xopizmene,
				v_new_idfpro,
				src.idfvknjrac,
				src.osnbezpod,
				src.datprometa,
				src.osnovicavt,
				src.osnovicant,
				src.prethpdvvt,
				src.prethpdvnt,
				src.tkonto,
				src.tsifra,
				src.pretpdv1vt,
				src.pretpdv1nt,
				src.pretpdv2vt,
				src.pretpdv2nt,
				src.irn,
				src.brojirn,
				src.demo,
				src.individualvatid,
				src.vatrecordingstatus,
				src.datum_stat_indvat,
				src.obj,
				src.beznak,
				src.epp_polje,
				src.fseppid
			FROM src
			RETURNING idkpr INTO v_new_idkpr;

			IF v_new_idkpr IS NOT NULL THEN
				UPDATE fpro
				SET idkpr = v_new_idkpr
				WHERE idfpro = v_new_idfpro;
			END IF;
		END IF;
	END LOOP;

	RETURN v_new_idfnal;
END;
$$;

CREATE OR REPLACE FUNCTION fn_kopiraj_stavke_u_postojeci_nalog(
	p_source_idfnal bigint,
	p_target_idfnal bigint,
	p_user text DEFAULT NULL
)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
	v_target fnal%ROWTYPE;
	v_source_exists boolean;
	v_next_rbr integer;
	v_new_idfpro bigint;
	v_new_idkir bigint;
	v_new_idkpr bigint;
	v_new_dug numeric;
	v_new_pot numeric;
	v_new_brst integer;
	r_fpro fpro%ROWTYPE;
BEGIN
	IF p_source_idfnal = p_target_idfnal THEN
		RAISE EXCEPTION 'Source and target FNAL must be different (idfnal=%)', p_source_idfnal;
	END IF;

	SELECT *
	INTO v_target
	FROM fnal
	WHERE idfnal = p_target_idfnal
	FOR UPDATE;

	IF NOT FOUND THEN
		RAISE EXCEPTION 'Target FNAL not found for idfnal=%', p_target_idfnal;
	END IF;

	SELECT EXISTS(
		SELECT 1
		FROM fnal
		WHERE idfnal = p_source_idfnal
	)
	INTO v_source_exists;

	IF NOT v_source_exists THEN
		RAISE EXCEPTION 'Source FNAL not found for idfnal=%', p_source_idfnal;
	END IF;

	SELECT COALESCE(MAX(fp.rbr), 0)
	INTO v_next_rbr
	FROM fpro fp
	WHERE fp.idfnal = p_target_idfnal;

	-- Copy FPRO rows one-by-one so we can attach new KIR/KPR rows per copied FPRO.
	FOR r_fpro IN
		SELECT *
		FROM fpro
		WHERE idfnal = p_source_idfnal
		ORDER BY rbr, idfpro
	LOOP
		v_next_rbr := v_next_rbr + 1;

		INSERT INTO fpro (
			god, kar, nalog, tipdok, rbr, danal, brst, iznos, kat, opis, dadok,
			rok, vrd, vkonta, konto, sifra, tra, deviznos, kurs, sifval, mi,
			ostatak, dokum, vkrbr, vktip, xdatunosa, xdatizmene, xopunos,
			xopizmene, idfnal, idfvknjrac, idorgjed, idfkpl, mag, komid, rdokid,
			mtroska, mestotrid, flegkomp, dokumv, dadokv, travez, starikonto,
			ojozn, sifkom, kolic, cena, idkir, idkpr
		)
		VALUES (
			v_target.god,
			v_target.kar,
			v_target.nalog,
			v_target.tipdok,
			v_next_rbr,
			v_target.danal,
			r_fpro.brst,
			r_fpro.iznos,
			r_fpro.kat,
			r_fpro.opis,
			r_fpro.dadok,
			r_fpro.rok,
			r_fpro.vrd,
			r_fpro.vkonta,
			r_fpro.konto,
			r_fpro.sifra,
			r_fpro.tra,
			r_fpro.deviznos,
			r_fpro.kurs,
			r_fpro.sifval,
			r_fpro.mi,
			r_fpro.ostatak,
			r_fpro.dokum,
			r_fpro.vkrbr,
			r_fpro.vktip,
			NOW(),
			NULL,
			COALESCE(NULLIF(p_user, ''), r_fpro.xopunos),
			r_fpro.xopizmene,
			p_target_idfnal,
			r_fpro.idfvknjrac,
			r_fpro.idorgjed,
			r_fpro.idfkpl,
			r_fpro.mag,
			r_fpro.komid,
			r_fpro.rdokid,
			r_fpro.mtroska,
			r_fpro.mestotrid,
			r_fpro.flegkomp,
			r_fpro.dokumv,
			r_fpro.dadokv,
			r_fpro.travez,
			r_fpro.starikonto,
			r_fpro.ojozn,
			r_fpro.sifkom,
			r_fpro.kolic,
			r_fpro.cena,
			NULL,
			NULL
		)
		RETURNING idfpro INTO v_new_idfpro;

		-- VRD=10 -> copy KIR and link copied FPRO.idkir.
		IF r_fpro.vrd = 10 THEN
			v_new_idkir := NULL;

			WITH src AS (
				SELECT *
				FROM kir k
				WHERE
					(r_fpro.idkir IS NOT NULL AND k.idkir = r_fpro.idkir)
					OR (r_fpro.idkir IS NULL AND k.idfpro = r_fpro.idfpro)
				ORDER BY k.idkir
				LIMIT 1
			)
			INSERT INTO kir (
				god, kar, vktip, vkrbr, krbr, idpartneri,
				dknjiz, danal, dizd, kracun,
				iznsapdv, oslobcl24, oslobcl25, izvozsapr, izvozbezpr,
				osn1, pdv1, osn2, pdv2, prom1, prom2,
				nalog, tipdok, vrd, dokum, vpr, rkar, brst,
				konto, sifra, pib, naziv,
				xdatunosa, xdatizmene, xopunos, xopizmene,
				idfvknjrac, idfpro, numdok, datprometa
			)
			SELECT
				v_target.god,
				v_target.kar,
				src.vktip,
				src.vkrbr,
				src.krbr,
				src.idpartneri,
				v_target.datob,
				v_target.danal,
				src.dizd,
				src.kracun,
				src.iznsapdv,
				src.oslobcl24,
				src.oslobcl25,
				src.izvozsapr,
				src.izvozbezpr,
				src.osn1,
				src.pdv1,
				src.osn2,
				src.pdv2,
				src.prom1,
				src.prom2,
				v_target.nalog,
				v_target.tipdok,
				src.vrd,
				src.dokum,
				src.vpr,
				src.rkar,
				src.brst,
				src.konto,
				src.sifra,
				src.pib,
				src.naziv,
				NOW(),
				NULL,
				COALESCE(NULLIF(p_user, ''), src.xopunos),
				src.xopizmene,
				src.idfvknjrac,
				v_new_idfpro,
				src.numdok,
				src.datprometa
			FROM src
			RETURNING idkir INTO v_new_idkir;

			IF v_new_idkir IS NOT NULL THEN
				UPDATE fpro
				SET idkir = v_new_idkir
				WHERE idfpro = v_new_idfpro;
			END IF;

		-- VRD=20 -> copy KPR and link copied FPRO.idkpr.
		ELSIF r_fpro.vrd = 20 THEN
			v_new_idkpr := NULL;

			WITH src AS (
				SELECT *
				FROM kpr k
				WHERE
					(r_fpro.idkpr IS NOT NULL AND k.idkpr = r_fpro.idkpr)
					OR (r_fpro.idkpr IS NULL AND k.idfpro = r_fpro.idfpro)
				ORDER BY k.idkpr
				LIMIT 1
			)
			INSERT INTO kpr (
				god, kar, vktip, vkrbr, drbr,
				dknjiz, duvoz, dizd,
				iznsapdv, iznoslob, nisuobvpdv, uvozbezpdv,
				prethpdv, pretpdv1, pretpdv2, uvozpdv,
				poljvred, poljpdv,
				vrd, konto, sifra, ter, uvozosnpdv,
				vpr, osnbezpdv, brst, rkar,
				nalog, dokum, pib, naziv, idpartneri,
				danal, tipdok,
				xdatunosa, xdatizmene, xopunos, xopizmene,
				idfpro, idfvknjrac,
				osnbezpod, datprometa,
				osnovicavt, osnovicant, prethpdvvt, prethpdvnt,
				tkonto, tsifra,
				pretpdv1vt, pretpdv1nt, pretpdv2vt, pretpdv2nt,
				irn, brojirn, demo, individualvatid, vatrecordingstatus,
				datum_stat_indvat, obj, beznak, epp_polje, fseppid
			)
			SELECT
				v_target.god,
				v_target.kar,
				src.vktip,
				src.vkrbr,
				src.drbr,
				v_target.datob,
				src.duvoz,
				src.dizd,
				src.iznsapdv,
				src.iznoslob,
				src.nisuobvpdv,
				src.uvozbezpdv,
				src.prethpdv,
				src.pretpdv1,
				src.pretpdv2,
				src.uvozpdv,
				src.poljvred,
				src.poljpdv,
				src.vrd,
				src.konto,
				src.sifra,
				src.ter,
				src.uvozosnpdv,
				src.vpr,
				src.osnbezpdv,
				src.brst,
				src.rkar,
				v_target.nalog,
				src.dokum,
				src.pib,
				src.naziv,
				src.idpartneri,
				v_target.danal,
				v_target.tipdok,
				NOW(),
				NULL,
				COALESCE(NULLIF(p_user, ''), src.xopunos),
				src.xopizmene,
				v_new_idfpro,
				src.idfvknjrac,
				src.osnbezpod,
				src.datprometa,
				src.osnovicavt,
				src.osnovicant,
				src.prethpdvvt,
				src.prethpdvnt,
				src.tkonto,
				src.tsifra,
				src.pretpdv1vt,
				src.pretpdv1nt,
				src.pretpdv2vt,
				src.pretpdv2nt,
				src.irn,
				src.brojirn,
				src.demo,
				src.individualvatid,
				src.vatrecordingstatus,
				src.datum_stat_indvat,
				src.obj,
				src.beznak,
				src.epp_polje,
				src.fseppid
			FROM src
			RETURNING idkpr INTO v_new_idkpr;

			IF v_new_idkpr IS NOT NULL THEN
				UPDATE fpro
				SET idkpr = v_new_idkpr
				WHERE idfpro = v_new_idfpro;
			END IF;
		END IF;
	END LOOP;

	SELECT
		COALESCE(SUM(CASE WHEN fp.kat IN (1, 2) THEN fp.iznos ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN fp.kat IN (3, 4) THEN fp.iznos ELSE 0 END), 0),
		COUNT(*)::integer
	INTO v_new_dug, v_new_pot, v_new_brst
	FROM fpro fp
	WHERE fp.idfnal = p_target_idfnal;

	UPDATE fnal
	SET
		dug = v_new_dug,
		pot = v_new_pot,
		brst = v_new_brst
	WHERE idfnal = p_target_idfnal;
END;
$$;
