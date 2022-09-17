/// This file is a part of MeetPlan Proton, which is a part of MeetPlanBackend (https://github.com/MeetPlan/MeetPlanBackend).
///
/// Copyright (c) 2022, Mitja Ševerkar <mytja@protonmail.com> and The MeetPlan Team.
/// All rights reserved.
/// Use of this source code is governed by the GNU AGPLv3 license, that can be found in the LICENSE file.

// Funkcije, v tej datoteki so zakomentirane, neuporabljene in tukaj shranjene samo za referenco.
// Izbriše se jih čez nekaj mesecev, s tem, da morajo biti v tem času narejeni minimalno trije commiti.

package proton

//func (p *protonImpl) FillGapsInTimetable(timetable []ProtonMeeting) ([]ProtonMeeting, error) {
//	protonMeetings := make([]ProtonMeeting, 0)
//	protonMeetings = append(protonMeetings, timetable...)
//
//	// Pojdimo čez vseh 5 šolskih dni (pon, tor, sre, čet in pet) in preverimo za luknje pri vsakemu razredu.
//	// Šolarji imajo generalno iste/podobne luknje, tako da ni potrebe po sestavljanju urnika za čisto vsakega učenca.
//	// Izbirni predmeti ipd. so tako ali tako predure ali po koncu generalnega pouka.
//
//	classes, err := p.db.GetClasses()
//	if err != nil {
//		return nil, err
//	}
//
//	subjectGroups := p.GetSubjectGroups()
//
//	for i := 0; i < len(classes); i++ {
//		depth := 0
//
//		class := classes[i]
//
//		var classStudents []string
//		err := json.Unmarshal([]byte(class.Students), &classStudents)
//		if err != nil {
//			return nil, err
//		}
//
//		constantHoleLen := 0
//		holeSame := 0
//
//		for {
//			timetable = make([]ProtonMeeting, 0)
//			timetable = append(timetable, protonMeetings...)
//
//			if depth >= (PROTON_ALLOWED_WHILE_DEPTH / 40) {
//				p.logger.Debug("exiting due to exceeded allowed depth")
//				return nil, errors.New("exceeded maximum allowed repeat depth")
//			}
//
//			var classTimetable = make([]ProtonMeeting, 0)
//
//			// 1. del
//			// Pridobimo vsa srečanja, ki so povezana z določenim razredom.
//			t, err := p.GetSubjectsOfClass(timetable, classStudents, class)
//			if err != nil {
//				return nil, err
//			}
//
//			// 2. del
//			// Zapolnimo luknje v urniku (z drugimi "bingljajočimi" urami) in ves čas preverjamo, če je vse v redu.
//			nonNormal := p.FindNonNormalHours(classTimetable)
//			holes := p.FindHoles(classTimetable)
//
//			if len(holes) == 0 || len(nonNormal) == 0 {
//				classTimetable, protonMeetings = p.PatchTheHoles(t, timetable)
//				break
//			}
//
//			if len(holes) == constantHoleLen {
//				holeSame++
//			} else {
//				p.logger.Debug("reset hole repeat counter")
//				holeSame = 0
//			}
//
//			constantHoleLen = len(holes)
//
//			if holeSame > 10 {
//				break
//			}
//
//			h := rand.Intn(len(holes))
//			n := rand.Intn(len(nonNormal))
//
//			hole := holes[h]
//			meeting := nonNormal[n]
//
//			var subjectGroup = make([]string, 0)
//
//			// TODO: Migriraj blok ure
//			for x := 0; x < len(subjectGroups); x++ {
//				group := subjectGroups[x]
//
//				var ok = false
//
//				for y := 0; y < len(group.Objects); y++ {
//					object := group.Objects[y]
//					if object.Type == "subject" && object.ObjectID == meeting.SubjectID {
//						ok = true
//					}
//				}
//
//				if !ok {
//					continue
//				}
//				for y := 0; y < len(group.Objects); y++ {
//					object := group.Objects[y]
//					if object.Type == "subject" && !helpers.Contains(subjectGroup, object.ObjectID) {
//						subjectGroup = append(subjectGroup, object.ObjectID)
//					}
//				}
//			}
//
//			if len(subjectGroup) == 0 {
//				subjectGroup = append(subjectGroup, meeting.SubjectID)
//			}
//
//			for y := 0; y < len(nonNormal); y++ {
//				m := nonNormal[y]
//				if m.Hour == meeting.Hour && m.DayOfTheWeek == meeting.DayOfTheWeek && helpers.Contains(subjectGroup, m.SubjectID) {
//					for x := 0; x < len(timetable); x++ {
//						meeting := timetable[x]
//						if meeting.ID == m.ID {
//							timetable = remove(timetable, x)
//						}
//					}
//
//					m.Hour = hole.Hour
//					m.DayOfTheWeek = hole.DayOfTheWeek
//					timetable = append(timetable, m)
//				}
//			}
//
//			for x := 0; x < len(timetable); x++ {
//				m := timetable[x]
//				if m.ID == meeting.ID {
//					timetable = remove(timetable, x)
//				}
//			}
//
//			meeting.Hour = hole.Hour
//			meeting.DayOfTheWeek = hole.DayOfTheWeek
//			timetable = append(timetable, meeting)
//
//			ok, err := p.CheckIfProtonConfigIsOk(timetable)
//			if ok {
//				p.logger.Debugw("successfully normalized proton timetable", "timetable", timetable, "protonMeetings", protonMeetings, "nonNormal", nonNormal, "classTimetable", classTimetable)
//
//				protonMeetings = make([]ProtonMeeting, 0)
//				protonMeetings = append(protonMeetings, timetable...)
//			} else {
//				p.logger.Debugw("failed to normalize proton-generated timetable", "error", err.Error(), "timetable", timetable, "protonTimetable", protonMeetings, "meeting", meeting)
//			}
//
//			depth++
//		}
//	}
//
//	p.logger.Debug("successfully normalized the timetable")
//	return protonMeetings, nil
//}
