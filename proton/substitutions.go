package proton

import "github.com/MeetPlan/MeetPlanBackend/helpers"

type TeacherTier struct {
	TeacherID   string
	Tier        int
	Name        string
	GradingList TierGradingList
}

type TierGradingList struct {
	TeacherID          string
	HasMeetingBefore   bool
	HasMeetingLater    bool
	HasMeeting2HBefore bool
	HasMeeting2HLater  bool
	TeachesSameSubject bool
	Name               string
}

func (p *protonImpl) ManageAbsences(meetingId string) ([]TeacherTier, error) {
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
	var preferredTeachers = make([]string, 0)
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
