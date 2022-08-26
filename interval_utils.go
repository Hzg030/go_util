package my_utils

type Interval struct {
	leftEqual     bool
	leftBoundary  *float64
	rightEqual    bool
	rightBoundary *float64
}

// InitInterval 构造方法
func InitInterval(leftBoundary *float64, leftEqual bool, rightBoundary *float64, rightEqual bool) *Interval {
	if leftBoundary == nil {
		leftEqual = false
	}
	if rightBoundary == nil {
		rightEqual = false
	}
	if *leftBoundary > *rightBoundary {
		//	TODO raise err?
		return nil
	}
	if *leftBoundary == *rightBoundary {
		leftEqual = true
		rightEqual = true
	}
	return &Interval{
		leftEqual:     leftEqual,
		leftBoundary:  leftBoundary,
		rightEqual:    rightEqual,
		rightBoundary: rightBoundary,
	}
}

// Intersect 取交集
func (i *Interval) Intersect(otherInterval *Interval) *Interval {
	// 不存在交集返回 nil
	if !i.Overlap(otherInterval) {
		return nil
	}
	var left *float64
	var right *float64
	var leftEqual bool
	var rightEqual bool

	if i.leftBoundary == nil || otherInterval.leftBoundary == nil {
		// 若其中一个为-∞， 取另一个作为左边界
		if i.leftBoundary == nil {
			left = otherInterval.leftBoundary
			leftEqual = otherInterval.leftEqual
		} else if otherInterval.leftBoundary == nil {
			left = i.leftBoundary
			leftEqual = i.leftEqual
		}
	} else {
		// 取较大值作为新的左边界
		if *i.leftBoundary < *otherInterval.leftBoundary {
			left = otherInterval.leftBoundary
			leftEqual = otherInterval.leftEqual
		} else if *i.leftBoundary == *otherInterval.leftBoundary {
			left = i.leftBoundary
			leftEqual = i.leftEqual
			if i.leftEqual == false || otherInterval.leftEqual == false {
				leftEqual = false
			}
		} else {
			left = i.leftBoundary
			leftEqual = i.leftEqual
		}
	}

	if i.rightBoundary == nil || otherInterval.rightBoundary == nil {
		// 若其中一个为+∞， 取另一个作为右边界
		if i.rightBoundary == nil {
			right = otherInterval.rightBoundary
			rightEqual = otherInterval.rightEqual
		} else if otherInterval.rightBoundary == nil {
			right = i.rightBoundary
			rightEqual = i.rightEqual
		}
	} else {
		// 取较小值作为新的右边界
		if *i.rightBoundary < *otherInterval.rightBoundary {
			right = i.rightBoundary
			rightEqual = i.rightEqual
		} else if *i.rightBoundary == *otherInterval.rightBoundary {
			right = i.rightBoundary
			rightEqual = i.leftEqual
			if i.rightEqual == false || otherInterval.rightEqual == false {
				rightEqual = false
			}
		} else {
			right = otherInterval.rightBoundary
			rightEqual = otherInterval.rightEqual
		}
	}
	return InitInterval(left, leftEqual, right, rightEqual)
}

// Union 取并集
func (i *Interval) Union(otherInterval *Interval) []*Interval {
	result := make([]*Interval, 0)
	// 不存在交集
	if !i.Overlap(otherInterval) {
		if *i.rightBoundary == *otherInterval.leftBoundary && (i.rightEqual == true || otherInterval.leftEqual == true) {
			result = append(result, InitInterval(i.leftBoundary, i.leftEqual, otherInterval.rightBoundary, otherInterval.rightEqual))
			return result
		}
		if *i.leftBoundary == *otherInterval.rightBoundary && (i.leftEqual == true || otherInterval.rightEqual == true) {
			result = append(result, InitInterval(otherInterval.leftBoundary, otherInterval.leftEqual, i.rightBoundary, i.rightEqual))
			return result
		}
		result = append(result, i)
		result = append(result, otherInterval)
		return result
	}
	var left *float64
	var leftEqual bool
	var right *float64
	var rightEqual bool

	if i.leftBoundary == nil || otherInterval.leftBoundary == nil {
		// 若其中一个为-∞， 取-∞作为左边界
		left = nil
		leftEqual = false
	} else {
		// 取较小值作为新的左边界
		if *i.leftBoundary < *otherInterval.leftBoundary {
			left = i.leftBoundary
			leftEqual = i.leftEqual
		} else {
			left = otherInterval.leftBoundary
			leftEqual = otherInterval.leftEqual
		}
	}

	if i.rightBoundary == nil || otherInterval.rightBoundary == nil {
		// 若其中一个为+∞， 取+∞作为右边界
		right = nil
		rightEqual = false
	} else {
		// 取较大值作为新的右边界
		if *i.rightBoundary < *otherInterval.rightBoundary {
			right = otherInterval.rightBoundary
			rightEqual = otherInterval.rightEqual
		} else {
			right = i.rightBoundary
			rightEqual = i.rightEqual
		}
	}
	result = append(result, InitInterval(left, leftEqual, right, rightEqual))
	return result
}

// Contains 判断是否包含另一个区间
func (i *Interval) Contains(otherInterval *Interval) bool {
	//不存在交集
	if !i.Overlap(otherInterval) {
		return false
	}
	if i.leftBoundary != nil && otherInterval.leftBoundary == nil {
		return false
	}
	if i.rightBoundary != nil && otherInterval.rightBoundary == nil {
		return false
	}
	if (i.leftBoundary != nil && otherInterval.leftBoundary != nil) && *i.leftBoundary > *otherInterval.leftBoundary {
		return false
	}
	if (i.rightBoundary != nil && otherInterval.rightBoundary != nil) && *i.rightBoundary < *otherInterval.rightBoundary {
		return false
	}
	return true
}

// Overlap 判断是否有交集
func (i *Interval) Overlap(otherInterval *Interval) bool {
	//不存在交集
	if (i.rightBoundary != nil && otherInterval.leftBoundary != nil) && ((*i.rightBoundary < *otherInterval.leftBoundary) ||
		(*i.rightBoundary == *otherInterval.leftBoundary && (i.rightEqual == false || otherInterval.leftEqual == false))) ||
		(i.leftBoundary != nil && otherInterval.rightBoundary != nil) && ((*i.leftBoundary > *otherInterval.rightBoundary) ||
			(*i.leftBoundary == *otherInterval.rightBoundary && (i.leftEqual == false || otherInterval.rightEqual == false))) {
		return false
	}
	return true
}
