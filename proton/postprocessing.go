// LEGACY
// needs updating to Proton v2

package proton

/*
package proton

import (
	"encoding/json"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
)
*/

const PROTON_MAX_NORMAL_HOUR = 8
const PROTON_MAX_AFTER_CLASS_HOUR = 8
const PROTON_REPEAT_POST_PROCESSING = 3

/*
// PostProcessHolesAndNonNormalHours poskrbi za bingljajoče ure.
//
// S funkcijo FindNonNormalHours poiščemo vse bingljajoče luknje in s temi urami poskušamo zapolniti, za zdaj samo, navadne luknje.
func (p *protonImpl) PostProcessHolesAndNonNormalHours(classTimetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting) {
	stableFullTimetable := make([]ProtonMeeting, 0)
	stableFullTimetable = append(stableFullTimetable, fullTimetable...)

	stableClassTimetable := make([]ProtonMeeting, 0)
	stableClassTimetable = append(stableClassTimetable, classTimetable...)

	holes := p.FindHoles(classTimetable)
	relationalHoles := p.FindRelationalHoles(classTimetable)

	nonNormal := p.FindNonNormalHours(classTimetable)

	for day := 0; day < len(holes); day++ {
		for hour := 0; hour < len(holes[day]); hour++ {
			hole := holes[day][hour]
			for i := 0; i < len(nonNormal); i++ {
				h := nonNormal[i]

				for n := 0; n < len(nonNormal); n++ {
					if !(h.Hour == nonNormal[n].Hour && h.DayOfTheWeek == nonNormal[n].DayOfTheWeek) {
						continue
					}

					for x := 0; x < len(fullTimetable); x++ {
						if fullTimetable[x].ID != nonNormal[n].ID {
							continue
						}

						fullTimetable[x].Hour = hole.Hour
						fullTimetable[x].DayOfTheWeek = hole.DayOfTheWeek
					}

					for x := 0; x < len(classTimetable); x++ {
						if classTimetable[x].ID != nonNormal[n].ID {
							continue
						}

						classTimetable[x].Hour = hole.Hour
						classTimetable[x].DayOfTheWeek = hole.DayOfTheWeek
					}
				}

				ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
				if ok {
					p.logger.Debugw("successfully moved the dangling hour to a hole", "hole", helpers.FmtSanitize(hole), "meeting", helpers.FmtSanitize(nonNormal[i]))

					stableFullTimetable = make([]ProtonMeeting, 0)
					stableFullTimetable = append(stableFullTimetable, fullTimetable...)

					stableClassTimetable = make([]ProtonMeeting, 0)
					stableClassTimetable = append(stableClassTimetable, classTimetable...)

					// Rekalkulirajmo luknje in nenormalne ure
					holes = p.FindHoles(classTimetable)
					nonNormal = p.FindNonNormalHours(classTimetable)

					// Zdaj smo zapolnili to luknjo in rekalkulirali luknje in reorganizirali srečanja, zato lahko zapustimo to zanko, kjer preverjamo, v katero luknjo bi kaj šlo.
					break
				} else {
					// V primeru, da se ne more dodeliti ta luknja, ni problema, gre samo na naslednje nenormalno srečanje v naslednji ponovni eksekuciji "for" zanke.

					p.logger.Debugw("failed while moving the dangling hour to a hole. reverting to stable state.", "err", err.Error(), "hole", helpers.FmtSanitize(hole), "meeting", helpers.FmtSanitize(nonNormal[i]))

					fullTimetable = make([]ProtonMeeting, 0)
					fullTimetable = append(fullTimetable, stableFullTimetable...)

					classTimetable = make([]ProtonMeeting, 0)
					classTimetable = append(classTimetable, stableClassTimetable...)
				}
			}
		}
	}

	for i := 0; i < len(relationalHoles); i++ {
		hole := relationalHoles[i]

		// To je inherited od prej
		for i := 0; i < len(nonNormal); i++ {
			h := nonNormal[i]

			for n := 0; n < len(nonNormal); n++ {
				if !(h.Hour == nonNormal[n].Hour && h.DayOfTheWeek == nonNormal[n].DayOfTheWeek) {
					continue
				}

				for x := 0; x < len(fullTimetable); x++ {
					if fullTimetable[x].ID != nonNormal[n].ID {
						continue
					}

					fullTimetable[x].Hour = hole.Hour
					fullTimetable[x].DayOfTheWeek = hole.DayOfTheWeek
				}

				for x := 0; x < len(classTimetable); x++ {
					if classTimetable[x].ID != nonNormal[n].ID {
						continue
					}

					classTimetable[x].Hour = hole.Hour
					classTimetable[x].DayOfTheWeek = hole.DayOfTheWeek
				}
			}

			ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
			if ok {
				p.logger.Debugw("successfully moved the dangling hour to a hole", "hole", helpers.FmtSanitize(hole), "meeting", helpers.FmtSanitize(nonNormal[i]))

				stableFullTimetable = make([]ProtonMeeting, 0)
				stableFullTimetable = append(stableFullTimetable, fullTimetable...)

				stableClassTimetable = make([]ProtonMeeting, 0)
				stableClassTimetable = append(stableClassTimetable, classTimetable...)

				// Rekalkulirajmo luknje in nenormalne ure
				holes = p.FindHoles(classTimetable)
				nonNormal = p.FindNonNormalHours(classTimetable)

				// Zdaj smo zapolnili to luknjo in rekalkulirali luknje in reorganizirali srečanja, zato lahko zapustimo to zanko, kjer preverjamo, v katero luknjo bi kaj šlo.
				break
			} else {
				// V primeru, da se ne more dodeliti ta luknja, ni problema, gre samo na naslednje nenormalno srečanje v naslednji ponovni eksekuciji "for" zanke.

				p.logger.Debugw("failed while moving the dangling hour to a hole. reverting to stable state.", "err", err.Error(), "hole", helpers.FmtSanitize(hole), "meeting", helpers.FmtSanitize(nonNormal[i]))

				fullTimetable = make([]ProtonMeeting, 0)
				fullTimetable = append(fullTimetable, stableFullTimetable...)

				classTimetable = make([]ProtonMeeting, 0)
				classTimetable = append(classTimetable, stableClassTimetable...)
			}
		}
	}

	return stableClassTimetable, stableFullTimetable
}

// PatchMistakes skrbi za popravljanje napak, ki se lahko zgodijo med post-procesiranjem urnika.
//
// Napake, katere ta funkcija odpravlja in popravlja:
//
//   - V enemu izmed tednov sta predmeta, ki bi morala biti v skupini srečanj, na popolnoma drugih lokacijah v urniku.
//     Primer:
//
//     TJA9a je 5. uro na ponedeljek,
//     TJA9b je 6. uro na petek
//     (obe srečanji sta v 2. tednu)
func (p *protonImpl) PatchMistakes(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting) {
	stableFullTimetable := make([]ProtonMeeting, 0)
	stableFullTimetable = append(stableFullTimetable, fullTimetable...)

	stableClassTimetable := make([]ProtonMeeting, 0)
	stableClassTimetable = append(stableClassTimetable, timetable...)

	weeks := OrderMeetingsByWeek(timetable)
	subjectGroups := p.GetSubjectGroups()

	subjects, err := p.db.GetAllSubjects()
	if err != nil {
		return stableClassTimetable, stableFullTimetable
	}

	subjectsNotInSubjectGroups := make([]string, 0)

	for i := 0; i < len(subjects); i++ {
		subject := subjects[i]

		//foundInSubjectGroup := false
		//
		//for n := 0; n < len(subjectGroups); n++ {
		//	subjectGroup := subjectGroups[n]
		//	for x := 0; x < len(subjectGroup.Objects); x++ {
		//		object := subjectGroup.Objects[x]
		//		if object.Type == "subject" && object.ObjectID == subject.ID {
		//			foundInSubjectGroup = true
		//			break
		//		}
		//	}
		//	if foundInSubjectGroup {
		//		break
		//	}
		//}
		//
		//if foundInSubjectGroup {
		//	continue
		//}

		subjectsNotInSubjectGroups = append(subjectsNotInSubjectGroups, subject.ID)
	}

	p.logger.Debugw("executing patching mistakes", "subjectsNotInSubjectGroups", helpers.FmtSanitize(subjectsNotInSubjectGroups))

	for i := 0; i < len(subjectGroups); i++ {
		subjectGroup := subjectGroups[i]
		subjects := make([]string, 0)
		for n := 0; n < len(subjectGroup.Objects); n++ {
			if subjectGroup.Objects[n].Type != "subject" {
				continue
			}
			subjects = append(subjects, subjectGroup.Objects[n].ObjectID)
		}
		for n := 0; n < len(weeks); n++ {
			week := weeks[n]

			stableWeek := make([]ProtonMeeting, 0)
			stableWeek = append(stableWeek, week...)

			for x := 0; x < len(week); x++ {
				meeting := week[x]

				if !helpers.Contains(subjects, meeting.SubjectID) {
					continue
				}

				// Zelo lepo ime, vem...
				isLonely := true
				for y := 0; y < len(week); y++ {
					m := week[y]
					if !helpers.Contains(subjects, m.SubjectID) || !(meeting.Hour == m.Hour && meeting.DayOfTheWeek == m.DayOfTheWeek) || meeting.ID == m.ID {
						continue
					}
					isLonely = false
					break
				}

				if !isLonely {
					continue
				}

				var lonelyHour *ProtonMeeting

				// Find other lonely hours
				for y := 0; y < len(week); y++ {
					m := week[y]
					if !helpers.Contains(subjects, m.SubjectID) || (meeting.Hour == m.Hour && meeting.DayOfTheWeek == m.DayOfTheWeek) || meeting.ID == m.ID {
						continue
					}

					isLonely := true

					// Naming go brrrrrr.
					for o := 0; o < len(week); o++ {
						m2 := week[o]
						if !helpers.Contains(subjects, m2.SubjectID) || !(m.Hour == m2.Hour && m.DayOfTheWeek == m2.DayOfTheWeek) || m.ID == m2.ID {
							continue
						}
						isLonely = false
						break
					}

					if !isLonely {
						continue
					}

					lonelyHour = &m
				}

				if lonelyHour == nil {
					// We haven't found the second lonely hour. Nothing we can do in this case.
					continue
				}

				lonelyMeeting := *lonelyHour

				p.logger.Debugw("found another lonely hour", "hour", helpers.FmtSanitize(meeting), "lonelyMeeting", helpers.FmtSanitize(lonelyMeeting))

				// Zdaj pa samo še damo eno uro na drugo (in seveda preverimo, če je vse kompatibilno s CheckIfProtonConfigIsOk funkcijo).
				newHour := lonelyMeeting.Hour
				newDay := lonelyMeeting.DayOfTheWeek
				if meeting.Hour < lonelyMeeting.Hour {
					newDay = meeting.DayOfTheWeek
					newHour = meeting.Hour
				}

				// Zamenjajmo v vseh živih seznamih (v go-ju se to imenuje "Slice")
				for y := 0; y < len(fullTimetable); y++ {
					if !(fullTimetable[y].ID == lonelyMeeting.ID || fullTimetable[y].ID == meeting.ID) {
						continue
					}
					fullTimetable[y].Hour = newHour
					fullTimetable[y].DayOfTheWeek = newDay
				}
				for y := 0; y < len(timetable); y++ {
					if !(timetable[y].ID == lonelyMeeting.ID || timetable[y].ID == meeting.ID) {
						continue
					}
					timetable[y].Hour = newHour
					timetable[y].DayOfTheWeek = newDay
				}
				for y := 0; y < len(week); y++ {
					if !(week[y].ID == lonelyMeeting.ID || week[y].ID == meeting.ID) {
						continue
					}
					week[y].Hour = newHour
					week[y].DayOfTheWeek = newDay
				}

				ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
				if ok {
					p.logger.Debugw("successfully moved lonely hours together",
						"lonelyMeeting", helpers.FmtSanitize(lonelyMeeting),
						"meeting", helpers.FmtSanitize(meeting),
						"newHour", helpers.FmtSanitize(newHour),
						"newDay", helpers.FmtSanitize(newDay),
					)

					stableFullTimetable = make([]ProtonMeeting, 0)
					stableFullTimetable = append(stableFullTimetable, fullTimetable...)

					stableClassTimetable = make([]ProtonMeeting, 0)
					stableClassTimetable = append(stableClassTimetable, timetable...)

					stableWeek = make([]ProtonMeeting, 0)
					stableWeek = append(stableWeek, week...)

					// Ponovno tedne
					weeks = OrderMeetingsByWeek(timetable)

					// Zdaj smo zapolnili to luknjo in rekalkulirali luknje in reorganizirali srečanja, zato lahko zapustimo to zanko, kjer preverjamo, v katero luknjo bi kaj šlo.
					continue
				}

				p.logger.Debugw(
					"failed while moving the lonely hours together. reverting to stable state.",
					"err", err.Error(),
					"lonelyMeeting", helpers.FmtSanitize(lonelyMeeting),
					"meeting", helpers.FmtSanitize(meeting),
					"newHour", helpers.FmtSanitize(newHour),
					"newDay", helpers.FmtSanitize(newDay),
				)

				fullTimetable = make([]ProtonMeeting, 0)
				fullTimetable = append(fullTimetable, stableFullTimetable...)

				timetable = make([]ProtonMeeting, 0)
				timetable = append(timetable, stableClassTimetable...)

				week = make([]ProtonMeeting, 0)
				week = append(week, stableWeek...)
			}
		}
	}

	// Posebej obravnavamo predmete, ki niso v skupinah predmetov
	// Bolj ali manj copy-pastano od zgoraj, ker sem len
	for i := 0; i < len(subjectsNotInSubjectGroups); i++ {
		for n := 0; n < len(weeks); n++ {
			week := weeks[n]

			stableWeek := make([]ProtonMeeting, 0)
			stableWeek = append(stableWeek, week...)

			for x := 0; x < len(week); x++ {
				meeting := week[x]

				if meeting.SubjectID != subjectsNotInSubjectGroups[i] {
					continue
				}

				if meeting.IsHalfHour {
					continue
				}

				// Zelo lepo ime, vem...
				isLonely := true
				for y := 0; y < len(weeks[1-n]); y++ {
					m := weeks[1-n][y]
					if m.SubjectID != meeting.SubjectID || !(meeting.Hour == m.Hour && meeting.DayOfTheWeek == m.DayOfTheWeek) || meeting.ID == m.ID {
						continue
					}
					isLonely = false
					break
				}

				if !isLonely {
					continue
				}

				var lonelyHour *ProtonMeeting

				// Find other lonely hours
				for y := 0; y < len(weeks[1-n]); y++ {
					m := weeks[1-n][y]

					if m.IsHalfHour {
						continue
					}

					//if meeting.SubjectID == 11 && m.SubjectID == meeting.SubjectID {
					//	p.logger.Debugw(
					//		"subject",
					//		"meeting", meeting,
					//		"isLonely", isLonely,
					//		"week", 1-n,
					//		"lonelyHour", m,
					//		"weekEquality", m.Week == meeting.Week,
					//		"subjectEquality", m.SubjectID != meeting.SubjectID,
					//		"hourDayEquality", meeting.Hour == m.Hour && meeting.DayOfTheWeek == m.DayOfTheWeek,
					//		"meetingIdEquality", meeting.ID == m.ID,
					//	)
					//}

					if m.SubjectID != meeting.SubjectID || (meeting.Hour == m.Hour && meeting.DayOfTheWeek == m.DayOfTheWeek) || meeting.ID == m.ID {
						continue
					}

					isLonely := true

					// Naming go brrrrrr.
					for o := 0; o < len(week); o++ {
						m2 := week[o]
						if m2.SubjectID != m.SubjectID || !(m.Hour == m2.Hour && m.DayOfTheWeek == m2.DayOfTheWeek) || m.ID == m2.ID {
							continue
						}
						p.logger.Debug(helpers.FmtSanitize(m), helpers.FmtSanitize(m2))
						isLonely = false
						break
					}

					if !isLonely {
						continue
					}

					lonelyHour = &m
				}

				if lonelyHour == nil {
					// We haven't found the second lonely hour. Nothing we can do in this case.
					continue
				}

				lonelyMeeting := *lonelyHour

				p.logger.Debugw("found another lonely hour for subject not in subject groups", "hour", helpers.FmtSanitize(meeting), "lonelyMeeting", helpers.FmtSanitize(lonelyMeeting), "subjectId", helpers.FmtSanitize(subjectsNotInSubjectGroups[i]))

				// Zdaj pa samo še damo eno uro na drugo (in seveda preverimo, če je vse kompatibilno s CheckIfProtonConfigIsOk funkcijo).
				newHour := lonelyMeeting.Hour
				newDay := lonelyMeeting.DayOfTheWeek
				if meeting.Hour < lonelyMeeting.Hour {
					newDay = meeting.DayOfTheWeek
					newHour = meeting.Hour
				}

				// Zamenjajmo v vseh živih seznamih (v go-ju se to imenuje "Slice")
				for y := 0; y < len(fullTimetable); y++ {
					if !(fullTimetable[y].ID == lonelyMeeting.ID || fullTimetable[y].ID == meeting.ID) {
						continue
					}
					fullTimetable[y].Hour = newHour
					fullTimetable[y].DayOfTheWeek = newDay
				}
				for y := 0; y < len(timetable); y++ {
					if !(timetable[y].ID == lonelyMeeting.ID || timetable[y].ID == meeting.ID) {
						continue
					}
					timetable[y].Hour = newHour
					timetable[y].DayOfTheWeek = newDay
				}
				for y := 0; y < len(week); y++ {
					if !(week[y].ID == lonelyMeeting.ID || week[y].ID == meeting.ID) {
						continue
					}
					week[y].Hour = newHour
					week[y].DayOfTheWeek = newDay
				}

				ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
				if ok {
					p.logger.Debugw("successfully moved lonely hours (w/out subject groups) together",
						"lonelyMeeting", helpers.FmtSanitize(lonelyMeeting),
						"meeting", helpers.FmtSanitize(meeting),
						"newHour", helpers.FmtSanitize(newHour),
						"newDay", helpers.FmtSanitize(newDay),
					)

					stableFullTimetable = make([]ProtonMeeting, 0)
					stableFullTimetable = append(stableFullTimetable, fullTimetable...)

					stableClassTimetable = make([]ProtonMeeting, 0)
					stableClassTimetable = append(stableClassTimetable, timetable...)

					stableWeek = make([]ProtonMeeting, 0)
					stableWeek = append(stableWeek, week...)

					// Ponovno tedne
					weeks = OrderMeetingsByWeek(timetable)

					// Zdaj smo zapolnili to luknjo in rekalkulirali luknje in reorganizirali srečanja, zato lahko zapustimo to zanko, kjer preverjamo, v katero luknjo bi kaj šlo.
					continue
				}

				p.logger.Debugw(
					"failed while moving the lonely hours (w/out subject groups) together. reverting to stable state.",
					"err", err.Error(),
					"lonelyMeeting", helpers.FmtSanitize(lonelyMeeting),
					"meeting", helpers.FmtSanitize(meeting),
					"newHour", helpers.FmtSanitize(newHour),
					"newDay", helpers.FmtSanitize(newDay),
				)

				fullTimetable = make([]ProtonMeeting, 0)
				fullTimetable = append(fullTimetable, stableFullTimetable...)

				timetable = make([]ProtonMeeting, 0)
				timetable = append(timetable, stableClassTimetable...)

				week = make([]ProtonMeeting, 0)
				week = append(week, stableWeek...)
			}
		}
	}

	return stableClassTimetable, stableFullTimetable
}

// PatchTheHoles skrbi za del krpanja lukenj pri normalnih predmetih in pourah (pri predurah ni potrebnega krpanja).
//
// PatchTheHoles je del post-procesirne suite funkcij.
func (p *protonImpl) PatchTheHoles(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting) {
	// Go naredi neke čudne stvari, podobno kot Python. Pri Python-u se, če bi shranil tole v novo spremenljivko (primer `a := b`)
	// bi se ustvaril pointer in bi bil vse samo en seznam. Tukaj je (iz nekega razloga) podobno, zato moramo ustvariti popolnoma nov seznam.
	stableFullTimetable := make([]ProtonMeeting, 0)
	stableFullTimetable = append(stableFullTimetable, fullTimetable...)

	stableClassTimetable := make([]ProtonMeeting, 0)
	stableClassTimetable = append(stableClassTimetable, timetable...)

	holes := p.FindHoles(timetable)

	orderedMeetings := OrderMeetingsByDay(timetable)

	beforeAfterSubjects := p.GetSubjectsBeforeOrAfterClass()

	for day := 0; day < 5; day++ {
		// Pojdimo čez vsak dan

		if len(holes[day]) == 0 {
			continue
		}

		var maxHour = -1

		for hour := 1; hour <= PROTON_MAX_NORMAL_HOUR+1; hour++ {
			// Predpostavljajmo, da se lahko zgenerira tudi blok ura na 8. uro.
			//
			// Ne mora se zgoditi, da se naše poure ustvarijo na 8. uro, temveč se lahko ustvarijo šele na 9. uro, zato predpostavljamo, da so vse ure v tem območju navadne.
			//
			// To je treba ves čas rekalkulirati.
			//
			// Delovanje te zanke:
			// Preštejmo vse (normalne) učne ure na ta dan za ta specifičen razred.
			// Tako lahko določimo najmanjšo možno uro za predmete po pouku (skrbimo, da se taki predmeti ne vrinejo med navadne učne ure (gl. pouk))

			meetingsHour := orderedMeetings[day][hour]
			if meetingsHour == nil || len(meetingsHour) == 0 {
				continue
			}

			found := true
			for _, value := range meetingsHour {
				if !helpers.Contains(beforeAfterSubjects, value.SubjectID) {
					found = false
					break
				}
			}

			if found {
				break
			}

			maxHour = hour
		}

		if maxHour == -1 {
			continue
		}

		for hour := 1; hour < len(orderedMeetings[day]); hour++ {
			// Pojdimo čez vsako uro.
			// Začnemo z ena, ker se tako ali tako predure ne štejejo.

			meetingsHour := orderedMeetings[day][hour]

			if meetingsHour == nil {
				continue
			}

			for h := 0; h < len(holes[day]); h++ {
				// Pojdimo čez vse luknje po vrstnem redu.

				hole := holes[day][h]
				if hole.Hour >= hour {
					// Luknje so urejene po vrstnem redu, tako da vemo, da ne bomo dobili manjšega števila in lahko samo preidemo na naslednje srečanje
					break
				}

				for m := 0; m < len(meetingsHour); m++ {
					// Pojdimo čez vsa srečanja v uri. Ta srečanja niso urejena v vrstnem redu, kar pa tudi ni pomembno.
					// Vsa srečanja, ki se prekrivajo, se bodo zamaknila na luknjo.
					// Obravnavamo samo srečanja DOLOČENEGA RAZREDA (glej `timetable` parameter funkcije) in ne celotnega urnika.

					meeting := meetingsHour[m]

					if meeting.IsHalfHour {
						p.logger.Debug("izognil sem se poluri", helpers.FmtSanitize(meeting))
						continue
					}

					if helpers.Contains(beforeAfterSubjects, meeting.SubjectID) && hole.Hour <= maxHour {
						// Preskočimo predure in poure (izbirne predmete) v primeru, da je luknja sredi normalnih predmetov
						continue
					}

					for x := 0; x < len(fullTimetable); x++ {
						// Zamenjamo uro v srečanju v nestabilnem fullTimetable-u.
						if meeting.ID == fullTimetable[x].ID {
							fullTimetable[x].Hour = hole.Hour
						}
					}
					for x := 0; x < len(timetable); x++ {
						// Zamenjamo uro v srečanju v nestabilnem fullTimetable-u.
						if meeting.ID == timetable[x].ID {
							timetable[x].Hour = hole.Hour
						}
					}
				}

				ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
				if ok {
					p.logger.Debugw("successfully patched the hole", "hole", helpers.FmtSanitize(hole), "meetings", helpers.FmtSanitize(meetingsHour))

					stableFullTimetable = make([]ProtonMeeting, 0)
					stableFullTimetable = append(stableFullTimetable, fullTimetable...)

					stableClassTimetable = make([]ProtonMeeting, 0)
					stableClassTimetable = append(stableClassTimetable, timetable...)

					// Rekalkulirajmo luknje in reorganizirajmo srečanja
					holes = p.FindHoles(timetable)
					orderedMeetings = OrderMeetingsByDay(timetable)

					// Rekalkulirajmo zadnjo normalno uro
					for hour := 1; hour <= PROTON_MAX_AFTER_CLASS_HOUR+1; hour++ {
						meetingsHour := orderedMeetings[day][hour]
						if meetingsHour == nil || len(meetingsHour) == 0 {
							continue
						}

						var ok = false

						for i := 0; i < len(meetingsHour); i++ {
							meeting := meetingsHour[i]
							if !helpers.Contains(beforeAfterSubjects, meeting.SubjectID) {
								ok = true
								break
							}
						}

						if !ok {
							continue
						}

						maxHour = hour
					}

					// Zdaj smo zapolnili to luknjo in rekalkulirali luknje in reorganizirali srečanja, zato lahko zapustimo to zanko, kjer preverjamo, v katero luknjo bi kaj šlo.
					break
				} else {
					// V primeru, da se ne more dodeliti ta luknja, ni problema, gre samo na naslednjo luknjo v naslednji ponovni eksekuciji "for" zanke.

					p.logger.Debugw("failed while patching the hole. reverting to stable state.", "err", err.Error(), "hole", helpers.FmtSanitize(hole), "meetings", helpers.FmtSanitize(meetingsHour))

					fullTimetable = make([]ProtonMeeting, 0)
					fullTimetable = append(fullTimetable, stableFullTimetable...)

					timetable = make([]ProtonMeeting, 0)
					timetable = append(timetable, stableClassTimetable...)
				}
			}
		}
	}

	return stableClassTimetable, stableFullTimetable
}

// SwapMeetings je eden izmed pomembnejših delov post-procesirne suite.
//
// Ta funkcija skrbi za menjanje ur.
//
// Navadno se zgodi, da so skupne ure (v skupinah srečanj) izrinjene, medtem ko ima en razred eno uro (za primer, recimo zgodovine), drug razred pa nima ure (tj. ima luknjo).
// Zgodi se, da so učenci (iz dveh razredov) zmešani med tema dvema predmetoma (v skupini srečanj) in so to edina nenormalna srečanja.
// V tem primeru se ne more zapolniti ta luknja, samo zaradi tega, ker ima en predmet zgodovino, kar pa ni optimalno za učence.
// Ta funkcija preveri za take ure in poskrbi za menjavo z vsemi predmeti, ki so v napoto na tisto uro in tako zapolni vse nepotrebne luknje.
//
// POMEMBNO: Funkcija ne menjava srečanj, ki niso v skupini srečanj, saj te navadno niso problematične.
// Če bi bile te ure problematične, bi lahko ustvarili skupino srečanj samo za en predmet.
func (p *protonImpl) SwapMeetings(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting) {
	stableFullTimetable := make([]ProtonMeeting, 0)
	stableFullTimetable = append(stableFullTimetable, fullTimetable...)

	stableClassTimetable := make([]ProtonMeeting, 0)
	stableClassTimetable = append(stableClassTimetable, timetable...)

	holes := p.FindHoles(timetable)
	nonNormal := p.FindNonNormalHours(timetable)

	subjectGroups := p.GetSubjectGroups()

	for i := 0; i < len(holes); i++ {
		holeDay := holes[i]
		for n := 0; n < len(holeDay); n++ {
			hole := holeDay[n]
			for x := 0; x < len(nonNormal); x++ {
				nonNormalHour := nonNormal[x]

				var currSubjectGroups = make([]ProtonRule, 0)

				for y := 0; y < len(subjectGroups); y++ {
					subjectGroup := subjectGroups[y]
					for o := 0; o < len(subjectGroup.Objects); o++ {
						object := subjectGroup.Objects[o]
						if object.Type == "subject" && object.ObjectID == nonNormalHour.SubjectID {
							currSubjectGroups = append(currSubjectGroups, subjectGroup)
							break
						}
					}
				}

				var subjectsInSubjectGroup = make([]string, 0)

				for y := 0; y < len(currSubjectGroups); y++ {
					subjectGroup := currSubjectGroups[y]
					for o := 0; o < len(subjectGroup.Objects); o++ {
						if subjectGroup.Objects[o].Type == "subject" && !helpers.Contains(subjectsInSubjectGroup, subjectGroup.Objects[o].ObjectID) {
							subjectsInSubjectGroup = append(subjectsInSubjectGroup, subjectGroup.Objects[o].ObjectID)
						}
					}
				}

				var replace1 = make([]ProtonMeeting, 0)
				var students1 = make([]string, 0)
				for y := 0; y < len(timetable); y++ {
					meeting := timetable[y]

					if !(meeting.Hour == nonNormalHour.Hour && meeting.DayOfTheWeek == nonNormalHour.DayOfTheWeek && helpers.Contains(subjectsInSubjectGroup, meeting.SubjectID)) {
						continue
					}

					if meeting.IsHalfHour {
						p.logger.Debug("izognil sem se poluri", helpers.FmtSanitize(meeting))
						continue
					}

					subject, err := p.db.GetSubject(meeting.SubjectID)
					if err != nil {
						return stableClassTimetable, stableFullTimetable
					}

					var s []string
					if subject.InheritsClass {
						class, err := p.db.GetClass(*subject.ClassID)
						if err != nil {
							return stableClassTimetable, stableFullTimetable
						}
						err = json.Unmarshal([]byte(class.Students), &s)
						if err != nil {
							return stableClassTimetable, stableFullTimetable
						}
					} else {
						err := json.Unmarshal([]byte(subject.Students), &s)
						if err != nil {
							return stableClassTimetable, stableFullTimetable
						}
					}

					for o := 0; o < len(s); o++ {
						student := s[o]
						if !helpers.Contains(students1, student) {
							students1 = append(students1, student)
						}
					}

					replace1 = append(replace1, meeting)
				}

				var replace2 = make([]ProtonMeeting, 0)
				for y := 0; y < len(fullTimetable); y++ {
					meeting := fullTimetable[y]
					if !(meeting.Hour == hole.Hour && meeting.DayOfTheWeek == hole.DayOfTheWeek) {
						continue
					}

					if meeting.IsHalfHour {
						p.logger.Debug("izognil sem se poluri", helpers.FmtSanitize(meeting))
						continue
					}

					subject, err := p.db.GetSubject(meeting.SubjectID)
					if err != nil {
						return stableClassTimetable, stableFullTimetable
					}

					var s []string
					if subject.InheritsClass {
						class, err := p.db.GetClass(*subject.ClassID)
						if err != nil {
							return stableClassTimetable, stableFullTimetable
						}
						err = json.Unmarshal([]byte(class.Students), &s)
						if err != nil {
							return stableClassTimetable, stableFullTimetable
						}
					} else {
						err := json.Unmarshal([]byte(subject.Students), &s)
						if err != nil {
							return stableClassTimetable, stableFullTimetable
						}
					}

					for o := 0; o < len(s); o++ {
						if helpers.Contains(students1, s[o]) {
							replace2 = append(replace2, meeting)
							break
						}
					}
				}

				// Zamenjajmo ure
				for y := 0; y < len(replace1); y++ {
					replace := replace1[y]
					for o := 0; o < len(timetable); o++ {
						if timetable[o].ID == replace.ID {
							timetable[o].Hour = hole.Hour
							timetable[o].DayOfTheWeek = hole.DayOfTheWeek
						}
					}
					for o := 0; o < len(fullTimetable); o++ {
						if fullTimetable[o].ID == replace.ID {
							fullTimetable[o].Hour = hole.Hour
							fullTimetable[o].DayOfTheWeek = hole.DayOfTheWeek
						}
					}
				}

				for y := 0; y < len(replace2); y++ {
					replace := replace2[y]
					for o := 0; o < len(timetable); o++ {
						if timetable[o].ID == replace.ID {
							timetable[o].Hour = nonNormalHour.Hour
							timetable[o].DayOfTheWeek = nonNormalHour.DayOfTheWeek
						}
					}
					for o := 0; o < len(fullTimetable); o++ {
						if fullTimetable[o].ID == replace.ID {
							fullTimetable[o].Hour = nonNormalHour.Hour
							fullTimetable[o].DayOfTheWeek = nonNormalHour.DayOfTheWeek
						}
					}
				}

				ok, err := p.CheckIfProtonConfigIsOk(fullTimetable)
				if ok {
					p.logger.Debugw(
						"successfully swapped the dangling hour with a hole",
						"hole", helpers.FmtSanitize(hole),
						"meeting", helpers.FmtSanitize(nonNormalHour),
						"subjectsInGroup", helpers.FmtSanitize(subjectsInSubjectGroup),
						"replace1", helpers.FmtSanitize(replace1),
						"replace2", helpers.FmtSanitize(replace2),
						"students1", helpers.FmtSanitize(students1),
						"subjectGroups", helpers.FmtSanitize(subjectGroups),
						"currSubjectGroups", helpers.FmtSanitize(currSubjectGroups),
					)

					stableFullTimetable = make([]ProtonMeeting, 0)
					stableFullTimetable = append(stableFullTimetable, fullTimetable...)

					stableClassTimetable = make([]ProtonMeeting, 0)
					stableClassTimetable = append(stableClassTimetable, timetable...)

					nonNormal = p.FindNonNormalHours(timetable)

					break
				} else {
					// V primeru, da se ne more dodeliti ta luknja, ni problema, gre samo na naslednje nenormalno srečanje v naslednji ponovni eksekuciji "for" zanke.

					p.logger.Debugw(
						"failed while swapping the dangling hour with a hole. reverting to stable state.",
						"err", err.Error(),
						"hole", helpers.FmtSanitize(hole),
						"meeting", helpers.FmtSanitize(nonNormalHour),
						"subjectsInGroup", helpers.FmtSanitize(subjectsInSubjectGroup),
						"replace1", helpers.FmtSanitize(replace1),
						"replace2", helpers.FmtSanitize(replace2),
						"students1", helpers.FmtSanitize(students1),
					)

					fullTimetable = make([]ProtonMeeting, 0)
					fullTimetable = append(fullTimetable, stableFullTimetable...)

					timetable = make([]ProtonMeeting, 0)
					timetable = append(timetable, stableClassTimetable...)
				}
			}
		}
	}

	return stableClassTimetable, stableFullTimetable
}

// TimetablePostProcessing skrbi za post-procesiranje urnika po koncu sestavljanja.
//
// Združuje naslednje funkcije post-procesiranja urnika:
//
// - PatchTheHoles (Stage 1), skrbi za polnjenje lukenj s premikanjem predmetov na prejšnje ure (predmeti se ne premikajo iz dneva na dan).
//
// - PostProcessHolesAndNonNormalHours (Stage 2), skrbi za polnjenje lukenj z nenormalnimi predmeti (bingljajočimi urami).
//
// - SwapMeetings (Stage 3), skrbi za menjavanje (bingljajočih) ur (v skupinah srečanj) in posledično masovno izboljšuje polnjenje lukenj.
//
// - PatchMistakes (Stage 4), skrbi za popravljanje napak pri generaciji in post-procesiranju urnika.
func (p *protonImpl) TimetablePostProcessing(stableTimetable []ProtonMeeting, class sql.Class, cancelPostProcessingBeforeDone bool) ([]ProtonMeeting, error) {
	var students []string
	err := json.Unmarshal([]byte(class.Students), &students)
	if err != nil {
		return nil, err
	}

	classTimetable, err := p.GetSubjectsOfClass(stableTimetable, students, class)
	if err != nil {
		return nil, err
	}

	if !p.FindIfHolesExist(classTimetable) && cancelPostProcessingBeforeDone {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 1
	classTimetable, stableTimetable = p.PatchTheHoles(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && cancelPostProcessingBeforeDone {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 2
	classTimetable, stableTimetable = p.PostProcessHolesAndNonNormalHours(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && cancelPostProcessingBeforeDone {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 3
	classTimetable, stableTimetable = p.SwapMeetings(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && cancelPostProcessingBeforeDone {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 4
	classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && cancelPostProcessingBeforeDone {
		return stableTimetable, nil
	}

	// Ponovi Stage 1
	classTimetable, stableTimetable = p.PatchTheHoles(classTimetable, stableTimetable)

	return stableTimetable, nil
}
*/
