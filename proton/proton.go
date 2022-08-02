/// This file is a part of MeetPlan Proton, which is a part of MeetPlanBackend (https://github.com/MeetPlan/MeetPlanBackend).
///
/// Copyright (c) 2022, Mitja Ševerkar <mytja@protonmail.com> and The MeetPlan Team.
/// All rights reserved.
/// Use of this source code is governed by the GNU AGPLv3 license, that can be found in the LICENSE file.

/// POZOR!
/// Ta package vsebuje kar nekaj matematike in nerazumljive kode.
/// Se ne priporoča brati, če ni komentarjev, saj bo tako najbolje za vaše mentalno zdravje.
/// Avtor ni odgovoren za kakršnekoli (materialne, fizične, mentalne ipd.) poškodbe med interakcijo s to kodo.

package proton

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/helpers"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"go.uber.org/zap"
	"time"
)

const PROTON_ALLOWED_WHILE_DEPTH = 1000
const PROTON_ALLOWED_FAIL_RATE = 500
const PROTON_ALLOWED_FAIL_RESET_RATE = 200
const PROTON_MAX_NORMAL_HOUR = 7
const PROTON_MIN_NORMAL_HOUR = 1
const PROTON_MAX_AFTER_CLASS_HOUR = 12
const PROTON_MIN_AFTER_CLASS_HOUR = 9
const PROTON_REPEAT_POST_PROCESSING = 3 // Večja je številka, večja je možnost, da nastane boljši urnik, a bo več časa trajalo, da se urnik post-procesira
const PROTON_CANCEL_POST_PROCESSING_BEFORE_DONE = false

//const PROTON_ALLOWED_HOLE_PATCHING_REPEAT_RATE = 200

type protonImpl struct {
	db     sql.SQL
	config ProtonConfig
	logger *zap.SugaredLogger
}

type Proton interface {
	ManageAbsences(meetingId int) ([]TeacherTier, error)

	NewProtonRule(rule ProtonRule) error
	GetProtonConfig() ProtonConfig

	// Suita funkcij, ki skrbijo za pravila

	GetAllRulesForTeacher(teacherId int) []ProtonRule
	GetSubjectGroups() []ProtonRule
	SubjectHasDoubleHours(subjectId int) bool
	CheckIfProtonConfigIsOk(timetable []ProtonMeeting) (bool, error)

	// FillGapsInTimetable(timetable []ProtonMeeting) ([]ProtonMeeting, error)

	// Post-procesirna suita funkcij

	TimetablePostProcessing(stableTimetable []ProtonMeeting, class sql.Class) ([]ProtonMeeting, error)

	PatchTheHoles(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
	GetSubjectsOfClass(timetable []ProtonMeeting, classStudents []int, class sql.Class) ([]ProtonMeeting, error)
	GetSubjectsBeforeOrAfterClass() []int
	GetSubjectsWithStackedHours() []int
	FindNonNormalHours(timetable []ProtonMeeting) []ProtonMeeting
	PostProcessHolesAndNonNormalHours(classTimetable []ProtonMeeting, stableTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
	FindHoles(timetable []ProtonMeeting) [][]ProtonMeeting
	FindRelationalHoles(timetable []ProtonMeeting) []ProtonMeeting
	SwapMeetings(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)
	PatchMistakes(timetable []ProtonMeeting, fullTimetable []ProtonMeeting) ([]ProtonMeeting, []ProtonMeeting)

	AssembleMeetingsFromProtonMeetings(timetable []ProtonMeeting, systemConfig sql.Config) ([]sql.Meeting, error)

	FindIfHolesExist(timetable []ProtonMeeting) bool

	SaveConfig(config ProtonConfig)
	DeleteRule(ruleId string)
}

func NewProton(db sql.SQL, logger *zap.SugaredLogger) (Proton, error) {
	protonConfig, err := LoadConfig()
	return &protonImpl{db: db, config: protonConfig, logger: logger}, err
}

type TierGradingList struct {
	TeacherID          int
	HasMeetingBefore   bool
	HasMeetingLater    bool
	HasMeeting2HBefore bool
	HasMeeting2HLater  bool
	TeachesSameSubject bool
	Name               string
}

type TeacherTier struct {
	TeacherID   int
	Tier        int
	Name        string
	GradingList TierGradingList
}

func (p *protonImpl) ManageAbsences(meetingId int) ([]TeacherTier, error) {
	teachers, err := p.db.GetTeachers()
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	originalMeeting, err := p.db.GetMeeting(meetingId)
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	subject, err := p.db.GetSubject(originalMeeting.SubjectID)
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	similarSubjects, err := p.db.GetSubjectsWithSpecificLongName(subject.LongName)
	if err != nil {
		return make([]TeacherTier, 0), err
	}
	var preferredTeachers = make([]int, 0)
	for i := 0; i < len(similarSubjects); i++ {
		subject := similarSubjects[i]
		if !helpers.Contains(preferredTeachers, subject.TeacherID) {
			preferredTeachers = append(preferredTeachers, subject.TeacherID)
		}
	}
	var teacherTiers = make([]TierGradingList, 0)
	for i := 0; i < len(teachers); i++ {
		teacher := teachers[i]
		teacherMeetings, err := p.db.GetMeetingsForTeacherOnSpecificDate(teacher.ID, originalMeeting.Date)
		if err != nil {
			return make([]TeacherTier, 0), err
		}
		var teacherTier = TierGradingList{
			TeacherID:          teacher.ID,
			HasMeetingBefore:   false,
			HasMeetingLater:    false,
			TeachesSameSubject: false,
			Name:               teacher.Name,
		}
		var hasSameHour = false

		// This should not be impacted by when the teacher has meetings or not
		if helpers.Contains(preferredTeachers, teacher.ID) {
			teacherTier.TeachesSameSubject = true
		}

		for n := 0; n < len(teacherMeetings); n++ {
			meeting := teacherMeetings[n]
			if meeting.Hour+1 == originalMeeting.Hour {
				teacherTier.HasMeetingBefore = true
			} else if meeting.Hour-1 == originalMeeting.Hour {
				teacherTier.HasMeetingLater = true
			} else if meeting.Hour == originalMeeting.Hour {
				hasSameHour = true
				break
			} else if meeting.Hour-2 == originalMeeting.Hour {
				teacherTier.HasMeeting2HLater = true
			} else if meeting.Hour+2 == originalMeeting.Hour {
				teacherTier.HasMeeting2HBefore = true
			}
		}
		if !hasSameHour {
			teacherTiers = append(teacherTiers, teacherTier)
		}
	}
	var recommendation = make([]TeacherTier, 0)
	for i := 0; i < len(teacherTiers); i++ {
		var tierGrade = 0
		teacherTier := teacherTiers[i]
		if teacherTier.TeachesSameSubject {
			tierGrade += 5
		}
		if teacherTier.HasMeeting2HLater {
			tierGrade += 1
		}
		if teacherTier.HasMeeting2HBefore {
			tierGrade += 1
		}
		if teacherTier.HasMeetingLater {
			tierGrade += 3
		}
		if teacherTier.HasMeetingBefore {
			tierGrade += 3
		}

		var skip = true

		for n := 0; n < len(recommendation); n++ {
			r := recommendation[n]
			if tierGrade > r.Tier {
				recommendation = helpers.Insert(recommendation, n, TeacherTier{
					TeacherID:   teacherTier.TeacherID,
					Tier:        tierGrade,
					Name:        teacherTier.Name,
					GradingList: teacherTier,
				})
				skip = false
				break
			}
		}

		if skip {
			recommendation = append(recommendation, TeacherTier{
				TeacherID:   teacherTier.TeacherID,
				Tier:        tierGrade,
				Name:        teacherTier.Name,
				GradingList: teacherTier,
			})
		}
	}
	return recommendation, err
}

func (p *protonImpl) GetAllRulesForTeacher(teacherId int) []ProtonRule {
	protonRules := make([]ProtonRule, 0)

	for i := 0; i < len(p.config.Rules); i++ {
		rule := p.config.Rules[i]
		for n := 0; n < len(rule.Objects); n++ {
			object := rule.Objects[n]
			if object.Type == "teacher" && object.ObjectID == teacherId {
				protonRules = append(protonRules, rule)
				break
			}
		}
	}

	return protonRules
}

func (p *protonImpl) GetSubjectGroups() []ProtonRule {
	protonRules := make([]ProtonRule, 0)
	for i := 0; i < len(p.config.Rules); i++ {
		protonRule := p.config.Rules[i]
		if protonRule.RuleType == 2 {
			protonRules = append(protonRules, protonRule)
		}
	}
	return protonRules
}

// SubjectHasDoubleHours preverja, če ima predmet blok ure.
func (p *protonImpl) SubjectHasDoubleHours(subjectId int) bool {
	for i := 0; i < len(p.config.Rules); i++ {
		rule := p.config.Rules[i]
		if rule.RuleType == 4 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && object.ObjectID == subjectId {
					return true
				}
			}
		}
	}
	return false
}

func (p *protonImpl) GetSubjectsOfClass(timetable []ProtonMeeting, classStudents []int, class sql.Class) ([]ProtonMeeting, error) {
	var classTimetable = make([]ProtonMeeting, 0)

	for n := 0; n < len(timetable); n++ {
		meeting := timetable[n]
		subject, err := p.db.GetSubject(meeting.SubjectID)
		if err != nil {
			return nil, err
		}
		if subject.InheritsClass && class.ID == subject.ClassID {
			classTimetable = append(classTimetable, meeting)
		} else {
			var students []int
			err := json.Unmarshal([]byte(subject.Students), &students)
			if err != nil {
				return nil, err
			}
			var ok = false
			for x := 0; x < len(students); x++ {
				subjectStudent := students[x]
				if helpers.Contains(classStudents, subjectStudent) {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}

			classTimetable = append(classTimetable, meeting)
		}
	}

	return classTimetable, nil
}

// GetSubjectsBeforeOrAfterClass retrieves all subjects, that are before or after the class (according to rule #3)
func (p *protonImpl) GetSubjectsBeforeOrAfterClass() []int {
	var subjects = make([]int, 0)
	rules := p.config.Rules
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if rule.RuleType == 3 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && !helpers.Contains(subjects, object.ObjectID) {
					subjects = append(subjects, object.ObjectID)
				}
			}
		}
	}
	return subjects
}

// GetSubjectsWithStackedHours retrieves all subjects, that have stacked hours (according to rule #4)
func (p *protonImpl) GetSubjectsWithStackedHours() []int {
	var subjects = make([]int, 0)
	rules := p.config.Rules
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		if rule.RuleType == 4 {
			for n := 0; n < len(rule.Objects); n++ {
				object := rule.Objects[n]
				if object.Type == "subject" && !helpers.Contains(subjects, object.ObjectID) {
					subjects = append(subjects, object.ObjectID)
				}
			}
		}
	}
	return subjects
}

func OrderMeetingsByDay(timetable []ProtonMeeting) [][][]ProtonMeeting {
	k := make([][][]ProtonMeeting, 5)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if meeting.Hour < 1 {
			continue
		}
		if k[meeting.DayOfTheWeek] == nil {
			// Večja številka, da ne dobimo overflowa
			k[meeting.DayOfTheWeek] = make([][]ProtonMeeting, 15)
		}
		if k[meeting.DayOfTheWeek][meeting.Hour] == nil {
			k[meeting.DayOfTheWeek][meeting.Hour] = make([]ProtonMeeting, 0)
		}
		k[meeting.DayOfTheWeek][meeting.Hour] = append(k[meeting.DayOfTheWeek][meeting.Hour], meeting)
	}

	return k
}

func (p *protonImpl) FindIfHolesExist(timetable []ProtonMeeting) bool {
	holes := p.FindHoles(timetable)
	for i := 0; i < len(holes); i++ {
		if len(holes[i]) != 0 {
			return true
		}
	}
	return false
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
						p.logger.Debug("izognil sem se poluri", meeting)
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
					p.logger.Debugw("successfully patched the hole", "hole", hole, "meetings", meetingsHour)

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

					p.logger.Debugw("failed while patching the hole. reverting to stable state.", "err", err.Error(), "hole", hole, "meetings", meetingsHour)

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

func (p *protonImpl) FindRelationalHoles(timetable []ProtonMeeting) []ProtonMeeting {
	var status = make(map[int]map[int][]ProtonMeeting)

	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if status[meeting.DayOfTheWeek] == nil {
			status[meeting.DayOfTheWeek] = make(map[int][]ProtonMeeting)
		}
		if status[meeting.DayOfTheWeek][meeting.Hour] == nil {
			status[meeting.DayOfTheWeek][meeting.Hour] = make([]ProtonMeeting, 0)
		}
		status[meeting.DayOfTheWeek][meeting.Hour] = append(status[meeting.DayOfTheWeek][meeting.Hour], meeting)
	}

	relationalHoles := make([]ProtonMeeting, 0)

	maxHour := 0
	minHour := -1

	// Iterate over each day
	for day := 0; day < 5; day++ {
		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}

		hCount := 0

		for hour := 1; hour <= PROTON_MAX_AFTER_CLASS_HOUR; hour++ {
			if status[day][hour] != nil {
				continue
			}

			hasHourAfter := false

			// Preverimo, če je po tej uri še kaj.
			// V primeru, da je, dodamo to kot luknjo
			// Zgodi se lahko, da se naša prva ura zgenerira na 12. uro, in če je blok ura, se zgenerira druga ura na 13. uro in je zato ne ujamemo.
			for n := hour + 1; n <= PROTON_MAX_AFTER_CLASS_HOUR+1; n++ {
				if status[day][n] != nil {
					hasHourAfter = true
					break
				}
			}

			if hasHourAfter {
				continue
			}

			hCount++
		}

		if hCount > maxHour {
			maxHour = hCount
		}
		if minHour == -1 || hCount < minHour {
			minHour = hCount
		}
	}

	// Preverimo za vse bingljajoče (neprave) luknje, ki nam niso ravno všeč.
	if minHour == -1 {
		return relationalHoles
	}

	relation := float32(minHour+maxHour) / 2
	for day := 0; day < 5; day++ {
		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}

		for hour := 1; hour <= int(relation); hour++ {
			if status[day][hour] != nil {
				continue
			}
			// Preverimo, če je po tej uri še kaj.
			// V primeru, da je, dodamo to kot luknjo
			// Zgodi se lahko, da se naša prva ura zgenerira na 12. uro, in če je blok ura, se zgenerira druga ura na 13. uro in je zato ne ujamemo.
			foundHourAfter := false
			for n := hour + 1; n <= 15; n++ {
				if status[day][n] != nil {
					foundHourAfter = true
					break
				}
			}

			if foundHourAfter {
				continue
			}

			relationalHoles = append(relationalHoles, ProtonMeeting{Hour: hour, DayOfTheWeek: day})
		}
	}

	p.logger.Debugw("found relational holes", "relationalHoles", relationalHoles)

	return relationalHoles
}

// FindHoles bo poiskal vse luknje vmes.
func (p *protonImpl) FindHoles(timetable []ProtonMeeting) [][]ProtonMeeting {
	// Ne me vprašat zakaj sem to naredil.
	var status = make(map[int]map[int][]ProtonMeeting)

	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if status[meeting.DayOfTheWeek] == nil {
			status[meeting.DayOfTheWeek] = make(map[int][]ProtonMeeting)
		}
		if status[meeting.DayOfTheWeek][meeting.Hour] == nil {
			status[meeting.DayOfTheWeek][meeting.Hour] = make([]ProtonMeeting, 0)
		}
		status[meeting.DayOfTheWeek][meeting.Hour] = append(status[meeting.DayOfTheWeek][meeting.Hour], meeting)
	}

	freeHours := make([][]ProtonMeeting, 5)

	// Iterate over each day
	for day := 0; day < 5; day++ {
		freeHours[day] = make([]ProtonMeeting, 0)

		if status[day] == nil {
			// Samo preskoči ta dan
			continue
		}

		for hour := 1; hour <= 15; hour++ {
			if status[day][hour] != nil {
				continue
			}
			// Preverimo, če je po tej uri še kaj.
			// V primeru, da je, dodamo to kot luknjo
			// Zgodi se lahko, da se naša prva ura zgenerira na 12. uro, in če je blok ura, se zgenerira druga ura na 13. uro in je zato ne ujamemo.
			for n := hour + 1; n <= PROTON_MAX_AFTER_CLASS_HOUR+1; n++ {
				if status[day][n] != nil {
					freeHours[day] = append(freeHours[day], ProtonMeeting{Hour: hour, DayOfTheWeek: day})
					break
				}
			}
		}
	}

	p.logger.Debugw("found holes", "freeHours", freeHours)

	return freeHours
}

// FindNonNormalHours poišče vse nenormalne ("bingljajoče") ure z dvema metodama:
//
// 1. metoda – najdi vse bingljajoče ure z izračunom povprečnega števila ur vsak dan skupaj. Ta metoda deluje večino časa, a pri ekstremih, ko je en dan praktično prazen, na žalost ne.
//
// 2. metoda – najdi vse bingljajoče ure z izračunom povprečja med dnevom z največ šolskimi urami in dnevom z najmanj šolskimi urami. To deluje, kadar 1. metoda ne deluje.
//
// Primer:
//
// Ponedeljek ima samo štiri (navadne – tj. niso izbirni predmeti) šolske ure,
// torek pa ima sedem (navadnih) šolskih ur.
// V tem primeru bo poskušal poiskati ustrezno rešitev, tako, da bo poiskal vse bingljajoče ure – v našem primeru šesto in sedmo šolsko uro v torku.
func (p *protonImpl) FindNonNormalHours(timetable []ProtonMeeting) []ProtonMeeting {
	beforeAfterSubjects := p.GetSubjectsBeforeOrAfterClass()
	days := make(map[int][]int)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if helpers.Contains(beforeAfterSubjects, meeting.SubjectID) {
			continue
		}
		if days[meeting.DayOfTheWeek] == nil {
			days[meeting.DayOfTheWeek] = make([]int, 0)
		}
		if !helpers.Contains(days[meeting.DayOfTheWeek], meeting.Hour) {
			days[meeting.DayOfTheWeek] = append(days[meeting.DayOfTheWeek], meeting.Hour)
		}
	}
	var totalHours = 0
	var totalDays = 0

	minHours := -1
	maxHours := -1

	for _, v := range days {
		if len(v) < minHours || minHours == -1 {
			minHours = len(v)
		}
		if len(v) > maxHours {
			maxHours = len(v)
		}
		totalHours += len(v)
		totalDays++
	}

	var meetings = make([]ProtonMeeting, 0)

	if minHours == -1 || maxHours == -1 {
		return meetings
	}

	average := float32((minHours + maxHours) / 2)

	if totalDays == 0 {
		return meetings
	}

	relation := float32(totalHours / totalDays)

	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if helpers.Contains(beforeAfterSubjects, meeting.SubjectID) {
			continue
		}
		if len(meetings) == 0 {
			meetings = append(meetings, meeting)
			continue
		}
		if meeting.Hour > int(relation) || meeting.Hour > int(average) {
			// Sortiramo po prioriteti (zadnje ure so vedno prve)
			for n := 0; n < len(meetings); n++ {
				if meetings[n].Hour < meeting.Hour {
					meetings = helpers.Insert(meetings, n, meeting)
					break
				}
			}
		}
	}

	p.logger.Debugw("found non-normal hours", "relation", relation, "totalHours", totalHours, "meetings", meetings, "totalDays", totalDays, "average", average)

	return meetings
}

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
					p.logger.Debugw("successfully moved the dangling hour to a hole", "hole", hole, "meeting", nonNormal[i])

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

					p.logger.Debugw("failed while moving the dangling hour to a hole. reverting to stable state.", "err", err.Error(), "hole", hole, "meeting", nonNormal[i])

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
				p.logger.Debugw("successfully moved the dangling hour to a hole", "hole", hole, "meeting", nonNormal[i])

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

				p.logger.Debugw("failed while moving the dangling hour to a hole. reverting to stable state.", "err", err.Error(), "hole", hole, "meeting", nonNormal[i])

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
// - V enemu izmed tednov sta predmeta, ki bi morala biti v skupini srečanj, na popolnoma drugih lokacijah v urniku.
//		Primer:
//
// 		TJA9a je 5. uro na ponedeljek,
// 		TJA9b je 6. uro na petek
// 		(obe srečanji sta v 2. tednu)
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

	subjectsNotInSubjectGroups := make([]int, 0)

	for i := 0; i < len(subjects); i++ {
		subject := subjects[i]

		foundInSubjectGroup := false

		for n := 0; n < len(subjectGroups); n++ {
			subjectGroup := subjectGroups[n]
			for x := 0; x < len(subjectGroup.Objects); x++ {
				object := subjectGroup.Objects[x]
				if object.Type == "subject" && object.ObjectID == subject.ID {
					foundInSubjectGroup = true
					break
				}
			}
			if foundInSubjectGroup {
				break
			}
		}

		if foundInSubjectGroup {
			continue
		}

		subjectsNotInSubjectGroups = append(subjectsNotInSubjectGroups, subject.ID)
	}

	p.logger.Debugw("executing patching mistakes", "subjectsNotInSubjectGroups", subjectsNotInSubjectGroups)

	for i := 0; i < len(subjectGroups); i++ {
		subjectGroup := subjectGroups[i]
		subjects := make([]int, 0)
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

				p.logger.Debugw("found another lonely hour", "hour", meeting, "lonelyMeeting", lonelyMeeting)

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
						"lonelyMeeting", lonelyMeeting,
						"meeting", meeting,
						"newHour", newHour,
						"newDay", newDay,
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
					"lonelyMeeting", lonelyMeeting,
					"meeting", meeting,
					"newHour", newHour,
					"newDay", newDay,
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

					if meeting.SubjectID == 11 && m.SubjectID == meeting.SubjectID {
						p.logger.Debugw(
							"subject",
							"meeting", meeting,
							"isLonely", isLonely,
							"week", 1-n,
							"lonelyHour", m,
							"weekEquality", m.Week == meeting.Week,
							"subjectEquality", m.SubjectID != meeting.SubjectID,
							"hourDayEquality", meeting.Hour == m.Hour && meeting.DayOfTheWeek == m.DayOfTheWeek,
							"meetingIdEquality", meeting.ID == m.ID,
						)
					}

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
						p.logger.Debug(m, m2)
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

				p.logger.Debugw("found another lonely hour for subject not in subject groups", "hour", meeting, "lonelyMeeting", lonelyMeeting, "subjectId", subjectsNotInSubjectGroups[i])

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
						"lonelyMeeting", lonelyMeeting,
						"meeting", meeting,
						"newHour", newHour,
						"newDay", newDay,
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
					"lonelyMeeting", lonelyMeeting,
					"meeting", meeting,
					"newHour", newHour,
					"newDay", newDay,
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
func (p *protonImpl) TimetablePostProcessing(stableTimetable []ProtonMeeting, class sql.Class) ([]ProtonMeeting, error) {
	var students []int
	err := json.Unmarshal([]byte(class.Students), &students)
	if err != nil {
		return nil, err
	}

	classTimetable, err := p.GetSubjectsOfClass(stableTimetable, students, class)
	if err != nil {
		return nil, err
	}

	if !p.FindIfHolesExist(classTimetable) && PROTON_CANCEL_POST_PROCESSING_BEFORE_DONE {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 1
	classTimetable, stableTimetable = p.PatchTheHoles(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && PROTON_CANCEL_POST_PROCESSING_BEFORE_DONE {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 2
	classTimetable, stableTimetable = p.PostProcessHolesAndNonNormalHours(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && PROTON_CANCEL_POST_PROCESSING_BEFORE_DONE {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 3
	classTimetable, stableTimetable = p.SwapMeetings(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && PROTON_CANCEL_POST_PROCESSING_BEFORE_DONE {
		classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)
		return stableTimetable, nil
	}

	// Stage 4
	classTimetable, stableTimetable = p.PatchMistakes(classTimetable, stableTimetable)

	if !p.FindIfHolesExist(classTimetable) && PROTON_CANCEL_POST_PROCESSING_BEFORE_DONE {
		return stableTimetable, nil
	}

	// Ponovi Stage 1
	classTimetable, stableTimetable = p.PatchTheHoles(classTimetable, stableTimetable)

	return stableTimetable, nil
}

type ProtonMeeting struct {
	Hour         int
	DayOfTheWeek int
	SubjectName  string
	SubjectID    int
	ID           string
	TeacherID    int
	Week         int
	ClassID      []int
	IsHalfHour   bool
}

// CheckIfProtonConfigIsOk preverja, če je trenuten timetable v redu sestavljen (v skladu z vsemi pravili).
// Ta funkcija je temelj vsega našega sistema.
func (p *protonImpl) CheckIfProtonConfigIsOk(timetable []ProtonMeeting) (bool, error) {
	// Predpriprava
	subjectGroups := p.GetSubjectGroups()

	// 1. korak
	// Pojdimo čez vse učitelje in preverimo, da se nič ne prekriva in je urnik skladen z učiteljevimi urami.
	teachers, err := p.db.GetTeachers()
	if err != nil {
		return false, err
	}
	for i := 0; i < len(teachers); i++ {
		teacher := teachers[i]
		for t := 0; t < len(timetable); t++ {
			// Pojdimo čez vse ure in preverimo, če se kaka ujema z učiteljem. Če se, nadaljujemo s postopkom.
			meeting := timetable[t]

			subject1, err := p.db.GetSubject(meeting.SubjectID)
			if err != nil {
				return false, err
			}

			if meeting.TeacherID != teacher.ID {
				continue
			}

			var subjectsInGroup = make([]int, 0)

			var hoursToday = 1

			// Preverimo, če se kake ure prekrivajo.
			for n := 0; n < len(timetable); n++ {
				meeting2 := timetable[n]

				if meeting2.ID == meeting.ID {
					continue
				}

				if meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting.SubjectID == meeting2.SubjectID && meeting.Week == meeting2.Week {
					hoursToday++
				}

				// Seveda moramo preveriti, če sta srečanji v isti proton skupini srečanj.
				var ok1 = false
				var ok2 = false
				for x := 0; x < len(subjectGroups); x++ {
					group := subjectGroups[x]
					for y := 0; y < len(group.Objects); y++ {
						if group.Objects[y].Type == "subject" && group.Objects[y].ObjectID == meeting.SubjectID {
							ok1 = true
							break
						}
					}
					if ok1 {
						for y := 0; y < len(group.Objects); y++ {
							if group.Objects[y].Type == "subject" {
								if meeting2.SubjectID == group.Objects[y].ObjectID {
									ok2 = true
								}
								if !helpers.Contains(subjectsInGroup, group.Objects[y].ObjectID) {
									subjectsInGroup = append(subjectsInGroup, group.Objects[y].ObjectID)
								}
							}
						}
					}
				}

				if meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting2.Hour == meeting.Hour && meeting.Week == meeting2.Week {
					// Preverimo, da se učencem ne prekriva

					subject2, err := p.db.GetSubject(meeting2.SubjectID)
					if err != nil {
						return false, err
					}
					if subject1.TeacherID == subject2.TeacherID {
						return false, errors.New("učitelj ne more učiti dveh predmetov ob istem času")
					}

					if subject1.InheritsClass && subject2.InheritsClass {
						if subject2.ClassID == subject1.ClassID {
							return false, errors.New(fmt.Sprintf("subjects %s and %s inherit the same class and thus cannot be made at same time", fmt.Sprint(subject1.ID), fmt.Sprint(subject2.ID)))
						}
						// V tem primeru nista isti razred, posledično se ne prekrivata
						continue
					}

					var students1 []int
					if subject1.InheritsClass {
						class, err := p.db.GetClass(subject1.ClassID)
						if err != nil {
							return false, err
						}
						err = json.Unmarshal([]byte(class.Students), &students1)
						if err != nil {
							return false, err
						}
					} else {
						err = json.Unmarshal([]byte(subject1.Students), &students1)
						if err != nil {
							return false, err
						}
					}

					var students2 []int
					if subject2.InheritsClass {
						class, err := p.db.GetClass(subject2.ClassID)
						if err != nil {
							return false, err
						}
						err = json.Unmarshal([]byte(class.Students), &students2)
						if err != nil {
							return false, err
						}
					} else {
						err = json.Unmarshal([]byte(subject2.Students), &students2)
						if err != nil {
							return false, err
						}
					}

					for s := 0; s < len(students1); s++ {
						student := students1[s]
						if helpers.Contains(students2, student) {
							return false, errors.New(fmt.Sprintf("subjects %s and %s contain the same student %s and thus cannot be made at same time", fmt.Sprint(subject1.ID), fmt.Sprint(subject2.ID), fmt.Sprint(student)))
						}
					}
				}

				if !(ok1 && ok2) && (meeting2.DayOfTheWeek == meeting.DayOfTheWeek && meeting2.Hour == meeting.Hour && meeting.Week == meeting2.Week) {
					return false, errors.New(
						fmt.Sprintf(
							"srečanji %s (%s %s) in pa %s (%s %s) se prekrivata - ne morem ustvariti urnika.",
							fmt.Sprint(meeting),
							fmt.Sprint(meeting.Hour),
							fmt.Sprint(meeting.DayOfTheWeek),
							fmt.Sprint(meeting2),
							fmt.Sprint(meeting2.Hour),
							fmt.Sprint(meeting2.DayOfTheWeek),
						),
					)
				}
			}

			if hoursToday > 2 {
				return false, errors.New("ne moreta biti več kot dve uri istega predmeta na en dan")
			}

			if len(subjectsInGroup) == 0 {
				subjectsInGroup = append(subjectsInGroup, meeting.SubjectID)
			}

			//p.logger.Debugw("subjects in group", "group", subjectsInGroup, "subjectGroups", subjectGroups, "meeting", meeting)

			mmap := make(map[int]bool)

			for s := 0; s < len(subjectsInGroup); s++ {
				subject, err := p.db.GetSubject(subjectsInGroup[s])
				if err != nil {
					return false, errors.New(fmt.Sprintf("subject %s is not found in the database - %s", fmt.Sprint(subjectsInGroup[s]), err.Error()))
				}
				rules := p.GetAllRulesForTeacher(subject.TeacherID)
				mmap[subjectsInGroup[s]] = false
				if len(rules) == 0 {
					mmap[subjectsInGroup[s]] = true
				}
				for r := 0; r < len(rules); r++ {
					rule := rules[r]
					if rule.RuleType == 0 {
						// Polni dnevi učitelja na šoli
						for n := 0; n < len(rule.Objects); n++ {
							object := rule.Objects[n]
							if object.Type == "day" && object.ObjectID == meeting.DayOfTheWeek {
								mmap[subjectsInGroup[s]] = true
								break
							}
						}
					} else if rule.RuleType == 1 {
						// Ure učitelja na šoli

						// Dan je treba izvleči posebej
						// TODO: Seznam dni namesto integerja
						day := -1
						for k := 0; k < len(rule.Objects); k++ {
							object := rule.Objects[k]
							if object.Type == "day" {
								day = object.ObjectID
							}
						}
						if day == -1 {
							return false, errors.New("neveljavno pravilo brez dni - pravilo št. 1")
						}

						if meeting.DayOfTheWeek != day {
							continue
						}

						for n := 0; n < len(rule.Objects); n++ {
							object := rule.Objects[n]

							if object.Type == "hour" && object.ObjectID == meeting.Hour {
								mmap[subjectsInGroup[s]] = true
								break
							}
						}
					}
				}
			}
			//fmt.Println(mmap, meeting)
			for n, v := range mmap {
				if !v {
					return false, errors.New(fmt.Sprintf("srečanje s predmetom %s se ne ujema s tem, kdaj je učitelj na šoli", fmt.Sprint(n)))
				}
			}
		}
	}
	return true, nil
}

func OrderMeetingsByWeek(timetable []ProtonMeeting) [][]ProtonMeeting {
	weeks := make([][]ProtonMeeting, 2)
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if weeks[meeting.Week] == nil {
			weeks[meeting.Week] = make([]ProtonMeeting, 0)
		}
		weeks[meeting.Week] = append(weeks[meeting.Week], meeting)
	}
	return weeks
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

				var subjectsInSubjectGroup = make([]int, 0)

				for y := 0; y < len(currSubjectGroups); y++ {
					subjectGroup := currSubjectGroups[y]
					for o := 0; o < len(subjectGroup.Objects); o++ {
						if subjectGroup.Objects[o].Type == "subject" && !helpers.Contains(subjectsInSubjectGroup, subjectGroup.Objects[o].ObjectID) {
							subjectsInSubjectGroup = append(subjectsInSubjectGroup, subjectGroup.Objects[o].ObjectID)
						}
					}
				}

				var replace1 = make([]ProtonMeeting, 0)
				var students1 = make([]int, 0)
				for y := 0; y < len(timetable); y++ {
					meeting := timetable[y]

					if !(meeting.Hour == nonNormalHour.Hour && meeting.DayOfTheWeek == nonNormalHour.DayOfTheWeek && helpers.Contains(subjectsInSubjectGroup, meeting.SubjectID)) {
						continue
					}

					if meeting.IsHalfHour {
						p.logger.Debug("izognil sem se poluri", meeting)
						continue
					}

					subject, err := p.db.GetSubject(meeting.SubjectID)
					if err != nil {
						return stableClassTimetable, stableFullTimetable
					}

					var s []int
					if subject.InheritsClass {
						class, err := p.db.GetClass(subject.ClassID)
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
						p.logger.Debug("izognil sem se poluri", meeting)
						continue
					}

					subject, err := p.db.GetSubject(meeting.SubjectID)
					if err != nil {
						return stableClassTimetable, stableFullTimetable
					}

					var s []int
					if subject.InheritsClass {
						class, err := p.db.GetClass(subject.ClassID)
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
						"hole", hole,
						"meeting", nonNormalHour,
						"subjectsInGroup", subjectsInSubjectGroup,
						"replace1", replace1,
						"replace2", replace2,
						"students1", students1,
						"subjectGroups", subjectGroups,
						"currSubjectGroups", currSubjectGroups,
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
						"hole", hole,
						"meeting", nonNormalHour,
						"subjectsInGroup", subjectsInSubjectGroup,
						"replace1", replace1,
						"replace2", replace2,
						"students1", students1,
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

func (p *protonImpl) AssembleMeetingsFromProtonMeetings(timetable []ProtonMeeting, systemConfig sql.Config) ([]sql.Meeting, error) {
	ok, err := p.CheckIfProtonConfigIsOk(timetable)
	if !ok {
		return nil, err
	}

	m := make(map[int]*time.Time)

	classes, err := p.db.GetClasses()
	if err != nil {
		return nil, err
	}

	lastSchoolDate := time.Unix(0, 0)

	// Preverimo za naše zadnje šolske dni in kateri je zadnji
	for i := 0; i < len(timetable); i++ {
		meeting := timetable[i]
		if m[meeting.SubjectID] == nil {
			subject, err := p.db.GetSubject(meeting.SubjectID)
			if err != nil {
				return nil, err
			}

			if subject.InheritsClass {
				class, err := p.db.GetClass(subject.ClassID)
				if err != nil {
					return nil, err
				}
				t := time.Unix(int64(class.LastSchoolDate), 0)
				m[meeting.SubjectID] = &t

				if t.After(lastSchoolDate) {
					lastSchoolDate = t
				}

				continue
			}

			var subjectStudents []int
			err = json.Unmarshal([]byte(subject.Students), &subjectStudents)
			if err != nil {
				return nil, err
			}

			for n := 0; n < len(classes); n++ {
				class := classes[n]

				var students []int
				err := json.Unmarshal([]byte(class.Students), &students)
				if err != nil {
					return nil, err
				}

				var ok = false
				for x := 0; x < len(students); x++ {
					if helpers.Contains(subjectStudents, students[x]) {
						ok = true
						break
					}
				}

				if !ok {
					continue
				}

				t := time.Unix(int64(class.LastSchoolDate), 0)

				if t.After(lastSchoolDate) {
					lastSchoolDate = t
				}

				if m[meeting.SubjectID] == nil || t.After(*m[meeting.SubjectID]) {
					m[meeting.SubjectID] = &t
				}
			}
		}
	}

	p.logger.Info("last school date", lastSchoolDate)

	weeks := OrderMeetingsByWeek(timetable)

	newTimetable := make([]sql.Meeting, 0)

	id := p.db.GetLastMeetingID()

	currentTime := time.Now()
	firstSchoolDay, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", fmt.Sprint(currentTime.Year()), "09", "01"))
	if err != nil {
		return nil, err
	}
	if firstSchoolDay.Weekday() == time.Sunday || firstSchoolDay.Weekday() == time.Saturday {
		// preskoči nekaj dni
		firstSchoolDay = firstSchoolDay.AddDate(0, 0, 1)
		if firstSchoolDay.Weekday() == time.Saturday {
			firstSchoolDay = firstSchoolDay.AddDate(0, 0, 1)
		}
	}
	firstMonday := firstSchoolDay.AddDate(0, 0, (-int(firstSchoolDay.Weekday()))+1)

	p.logger.Debugw(
		"calculated first school day",
		"schoolDay", firstSchoolDay,
		"currentTime", currentTime,
		"firstMonday", firstMonday,
		"firstSchoolWeekday", int(firstSchoolDay.Weekday()),
	)

	firstWeek := true
	weekCount := 0

	for {
		triggeredEndOfSchool := false

		for w := 0; w < len(weeks); w++ {
			week := weeks[w]

			for i := 0; i < len(week); i++ {
				meeting := week[i]

				if firstWeek && meeting.DayOfTheWeek < (int(firstSchoolDay.Weekday())-1) {
					continue
				}

				date := firstMonday.AddDate(0, 0, weekCount*7+meeting.DayOfTheWeek)
				p.logger.Debug(date)

				if (date.Day() == lastSchoolDate.Day() && date.Month() == lastSchoolDate.Month() && date.Year() == lastSchoolDate.Year()) || date.After(lastSchoolDate) {
					p.logger.Info("triggered last school date")
					triggeredEndOfSchool = true
				}

				nowDate := date.Format("02-01-2006")

				vacationDate := date.Format("2006-01-02")

				if helpers.Contains(systemConfig.SchoolFreeDays, vacationDate) {
					p.logger.Debugw("skipped meeting due to vacation", "date", vacationDate, "meeting", meeting)
					continue
				}

				subject, err := p.db.GetSubject(meeting.SubjectID)
				if err != nil {
					return nil, err
				}

				newTimetable = append(newTimetable, sql.Meeting{
					ID:                  id,
					MeetingName:         meeting.SubjectName,
					TeacherID:           meeting.TeacherID,
					SubjectID:           meeting.SubjectID,
					Hour:                meeting.Hour,
					Location:            subject.Location,
					Date:                nowDate,
					IsMandatory:         true,
					URL:                 "",
					Details:             "",
					IsSubstitution:      false,
					IsGrading:           false,
					IsWrittenAssessment: false,
					IsTest:              false,
					IsBeta:              true,
				})
				id++
			}
			firstWeek = false
			weekCount++

			if triggeredEndOfSchool {
				break
			}
		}

		if triggeredEndOfSchool {
			p.logger.Debug("triggered end of school")
			break
		}
	}

	return newTimetable, nil
}

func (p *protonImpl) NewProtonRule(rule ProtonRule) error {
	config, err := AddNewRule(p.config, rule)
	if err != nil {
		return err
	}
	p.config = config
	return nil
}

func (p *protonImpl) SaveConfig(config ProtonConfig) {
	err := SaveConfig(config)
	if err != nil {
		return
	}
	p.config = config
}

func (p *protonImpl) DeleteRule(ruleId string) {
	for i := 0; i < len(p.config.Rules); i++ {
		if p.config.Rules[i].ID == ruleId {
			p.config.Rules = helpers.Remove(p.config.Rules, i)
			SaveConfig(p.config)
			return
		}
	}
}

func (p *protonImpl) GetProtonConfig() ProtonConfig {
	return p.config
}
