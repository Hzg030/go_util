package scope_filter

import (
	"reflect"

	"git.garena.com/people/core-base/hris-field-management/model/errorcode"
)

type MDBitMap struct {
	lengthList []int64
	mapValue   []interface{}
}

// InitMDBitMap 构造方法
/**
 * @Author zenggui.huang
 * @Description 初始化 MDBitMap 多维位图
 * @Date 6:06 下午 2022/8/25
 * @Param lengthList MDBitMap 各维度长度， indexList 设置为true的下标数组
 * @return
 * @e.g.
	输入：length: [3,4], indexList : [[0,1], [2,3]]
	返回一个 3*4 的二维位图，并将 [0,1], [2,3] 下标元素初始化为true:
	0 1 0 0
	0 0 0 0
	0 0 0 1
 **/
func (m *MDBitMap) InitMDBitMap(lengthList []int64, indexList [][]int64) error {
	if len(lengthList) == 0 {
		return errorcode.LengthListEmpty
	}
	//任一维度的 长度不能<=0
	for _, length := range lengthList {
		if length <= 0 {
			return errorcode.DimensionLengthTooSmall
		}
	}
	m.lengthList = lengthList
	m.createEmptyMDBitmap()
	if indexList != nil && len(indexList) > 0 {
		err := m.initMDBitMapByIndexList(indexList)
		if err != nil {
			return err
		}
	}
	return nil
}

//初始化空多维位图
//示例输入：lengthList = {3,4,5,6}
//则输出一个 [3][4][5][6]bool 多维位图，其中所有取值都是false
func (m *MDBitMap) createEmptyMDBitmap() {
	var mDBitMap []interface{}
	//从内往外 创建 []interface{} ， 并一层层封装
	for i := len(m.lengthList) - 1; i >= 0; i-- {
		if mDBitMap == nil {
			mDBitMap = make([]interface{}, m.lengthList[i])
			for j := 0; j < int(m.lengthList[i]); j++ {
				mDBitMap[j] = false
			}
			continue
		}
		tempBit := make([]interface{}, m.lengthList[i])
		for j := 0; j < int(m.lengthList[i]); j++ {
			copyBit := deepCopy(mDBitMap)
			tempBit[j] = copyBit
		}
		mDBitMap = tempBit
	}
	m.mapValue = mDBitMap
}

//根据上送的 下标列表 indexList；将对应下标元素设置为true
func (m *MDBitMap) initMDBitMapByIndexList(indexList [][]int64) error {
	for _, subIndexList := range indexList {
		//subIndexList 长度必须与lengthList长度一致，否则无法填充
		if len(subIndexList) != len(m.lengthList) {
			return errorcode.InconsistentLength
		}
		var tempBit []interface{}
		for i, index := range subIndexList {
			//index 下标范围必须在 lengthList[i] 中
			if index < 0 || index >= m.lengthList[i] {
				return errorcode.MDBitMapIndexOutOfRange
			}
			if i == len(subIndexList)-1 {
				break
			}
			if tempBit == nil {
				tempBit = m.mapValue[index].([]interface{})
				continue
			}
			tempBit = tempBit[index].([]interface{})
		}
		tempBit[subIndexList[len(subIndexList)-1]] = true
	}
	return nil
}

/*
获取当前 MDBitMap 所有下标列表
假设当前 MDBitMap 的lengthList = [3,4]
则返回 allIndexList = [
		[0,0],[0,1],[0,2],[0,3],
		[1,0],[1,1],[1,2],[1,3],
		[2,0],[2,1],[2,2],[2,3],
	  ]
*/
func (m *MDBitMap) getAllIndexList() [][]int64 {
	allIndexList := make([][]int64, 0)
	for i := len(m.lengthList) - 1; i >= 0; i-- {
		subIndexList := make([][]int64, 0)
		if len(allIndexList) == 0 {
			for j := 0; j < int(m.lengthList[i]); j++ {
				temp := make([]int64, len(m.lengthList))
				temp[i] = int64(j)
				subIndexList = append(subIndexList, temp)
			}
			allIndexList = subIndexList
			continue
		}
		for j := 0; j < len(allIndexList); j++ {
			for k := 0; k < int(m.lengthList[i]); k++ {
				copyList := make([]int64, len(m.lengthList))
				copy(copyList, allIndexList[j])
				copyList[i] = int64(k)
				subIndexList = append(subIndexList, copyList)
			}
		}
		allIndexList = subIndexList
	}
	return allIndexList
}

// 	OrMDBitMap 或运算， 与targetMDBitMap 位图做或运算并返回新的 bitMap
func (m *MDBitMap) OrMDBitMap(targetBitMap *MDBitMap) (*MDBitMap, error) {
	// 若结构不一致，抛出异常
	if !reflect.DeepEqual(m.lengthList, targetBitMap.lengthList) {
		return nil, errorcode.InconsistentMap
	}
	finalMDBitMap := MDBitMap{}
	finalMDBitMap.InitMDBitMap(m.lengthList, nil)
	//获取所有下标 allIndexList 以进行逐个遍历运算
	allIndexList := m.getAllIndexList()
	for _, indexList := range allIndexList {
		var source []interface{}
		var target []interface{}
		var final []interface{}
		for i, index := range indexList {
			if i == len(indexList)-1 {
				break
			}
			if i == 0 {
				source = m.mapValue[index].([]interface{})
				target = targetBitMap.mapValue[index].([]interface{})
				final = finalMDBitMap.mapValue[index].([]interface{})
			} else {
				source = source[index].([]interface{})
				target = target[index].([]interface{})
				final = final[index].([]interface{})
			}
		}
		final[indexList[len(indexList)-1]] = source[indexList[len(indexList)-1]].(bool) || target[indexList[len(indexList)-1]].(bool)
	}
	return &finalMDBitMap, nil
}

// 	AndMDBitMap 与运算， 与targetMDBitMap 位图做与运算并返回新的 bitMap
func (m *MDBitMap) AndMDBitMap(targetBitMap *MDBitMap) (*MDBitMap, error) {
	// 若结构不一致，抛出异常
	if !reflect.DeepEqual(m.lengthList, targetBitMap.lengthList) {
		return nil, errorcode.InconsistentMap
	}
	finalMDBitMap := MDBitMap{}
	finalMDBitMap.InitMDBitMap(m.lengthList, nil)
	//获取所有下标 allIndexList 以进行逐个遍历运算
	allIndexList := m.getAllIndexList()
	for _, indexList := range allIndexList {
		var source []interface{}
		var target []interface{}
		var final []interface{}
		for i, index := range indexList {
			if i == len(indexList)-1 {
				break
			}
			if i == 0 {
				source = m.mapValue[index].([]interface{})
				target = targetBitMap.mapValue[index].([]interface{})
				final = finalMDBitMap.mapValue[index].([]interface{})
			} else {
				source = source[index].([]interface{})
				target = target[index].([]interface{})
				final = final[index].([]interface{})
			}
		}
		final[indexList[len(indexList)-1]] = source[indexList[len(indexList)-1]].(bool) && target[indexList[len(indexList)-1]].(bool)
	}
	return &finalMDBitMap, nil
}

// 	NotMDBitMap 取反运算， 将当前位图做取反运算并返回新的 bitMap
func (m *MDBitMap) NotMDBitMap() *MDBitMap {
	finalMDBitMap := MDBitMap{}
	finalMDBitMap.InitMDBitMap(m.lengthList, nil)
	//获取所有下标 allIndexList 以进行逐个遍历运算
	allIndexList := m.getAllIndexList()
	for _, indexList := range allIndexList {
		var source []interface{}
		var final []interface{}
		for i, index := range indexList {
			if i == len(indexList)-1 {
				break
			}
			if i == 0 {
				source = m.mapValue[index].([]interface{})
				final = finalMDBitMap.mapValue[index].([]interface{})
			} else {
				source = source[index].([]interface{})
				final = final[index].([]interface{})
			}
		}
		final[indexList[len(indexList)-1]] = !source[indexList[len(indexList)-1]].(bool)
	}
	return &finalMDBitMap
}

// EqualMDBitMap 判断与另一个 MDBitMap 是否相等
func (m *MDBitMap) EqualMDBitMap(targetBitMap *MDBitMap) bool {
	return reflect.DeepEqual(m, targetBitMap)
}

//	ContainsMDBitMap 判断是否包含另一个 MDBitMap
//	若 A&B = B ，说明 A 包含 B
func (m *MDBitMap) ContainsMDBitMap(targetBitMap *MDBitMap) (bool, error) {
	temp, err := m.AndMDBitMap(targetBitMap)
	if err != nil {
		return false, err
	}
	return targetBitMap.EqualMDBitMap(temp), nil
}

//构建初始化下标数组
//假设 是四维位图(每个维度长度为10)，其中第三维 function 是 role,其取值范围是[admin]，下标取值为1
//则 该四维位图中 [1-9][1-9][1][1-9] 下标的值都取true
/* 可参考二维例子 ，
示例： or:{A in [A1]}  即[A1][B0], [A1][B1], [A1][B2], [A1][B3]都为 true
A/B B0 B1 B2 B3
A0  0  1  1  0
A1  0  1  1  0
A2  0  1  1  0
A3  0  1  1  0
则该方法输出为 设置为true的下标数组列表
示例： [[1,0], [1,1], [1,2], [1,3]]
*/
func getBitMapIndexList(function string, dataScopeList []interface{}, functionValueIndexMap map[string]map[interface{}]int64, rangeFunctionValueIndexMap map[string]map[interface{}]int64, functionIndexMap map[string]int64) [][]int64 {
	finalBitMapIndexList := make([][]int64, 0)
	// 遍历function data_scope 取值
	for _, value := range dataScopeList {
		valueBitMapIndexList := make([][]int64, 0)
		for otherFunction, otherFuncDataScope := range functionValueIndexMap {
			if otherFunction == function {
				continue
			}
			//valueOrBitMapIndexList为空，进行初始化
			//or:{A in [A1]} valueBitMapIndexList = [[1,0], [1,1], [1,2], [1,3]]
			if len(valueBitMapIndexList) == 0 {
				for i := 0; i < len(otherFuncDataScope); i++ {
					valueBitMapIndexList = append(valueBitMapIndexList, make([]int64, len(functionIndexMap)))
					valueBitMapIndexList[i][functionIndexMap[function]] = functionValueIndexMap[function][value]
					valueBitMapIndexList[i][functionIndexMap[otherFunction]] = int64(i)
				}
				continue
			}
			//若还有其他维度，则需要在valueOrBitMapIndexList的基础上继续填充
			/*示例 三维位图 A-B-C
			条件为  or:{A in [A1]}
			则C 维度的[[1,0], [1,1], [1,2], [1,3]]都需要填充
			也就是
			[1,0,0], [1,1,0], [1,2,0], [1,3,0]
			[1,0,1], [1,1,1], [1,2,1], [1,3,1]
			[1,0,2], [1,1,2], [1,2,2], [1,3,2]
				 C	   C1               C2               C3
				A/B B0 B1 B2 B3  A/B B0 B1 B2 B3  A/B B0 B1 B2 B3
				A0  0  0  0  0   A0  0  0  0  0   A0  0  0  0  0
				A1  1  1  1  1   A1  1  1  1  1   A1  1  1  1  1
				A2  0  0  0  0   A2  0  0  0  0   A2  0  0  0  0
				A3  0  0  0  0   A3  0  0  0  0   A3  0  0  0  0
			*/
			tempIndexList := make([][]int64, 0)
			for i := 0; i < len(valueBitMapIndexList); i++ {
				for j := 0; j < len(otherFuncDataScope); j++ {
					copyList := make([]int64, len(functionIndexMap))
					copy(copyList, valueBitMapIndexList[i])
					copyList[functionIndexMap[function]] = functionValueIndexMap[function][value]
					copyList[functionIndexMap[otherFunction]] = int64(j)
					tempIndexList = append(tempIndexList, copyList)
				}
			}
			valueBitMapIndexList = tempIndexList
		}
		//处理范围值列表 rangeFunctionValueIndexMap
		for otherFunction, otherFuncDataScope := range rangeFunctionValueIndexMap {
			if otherFunction == function {
				continue
			}
			//valueOrBitMapIndexList为空，进行初始化
			//or:{A in [A1]} valueBitMapIndexList = [[1,0], [1,1], [1,2], [1,3]]
			if len(valueBitMapIndexList) == 0 {
				for i := 0; i < 2*len(otherFuncDataScope)+1; i++ {
					valueBitMapIndexList = append(valueBitMapIndexList, make([]int64, len(functionIndexMap)))
					valueBitMapIndexList[i][functionIndexMap[function]] = functionValueIndexMap[function][value]
					valueBitMapIndexList[i][functionIndexMap[otherFunction]] = int64(i)
				}
				continue
			}
			//若还有其他维度，则需要在valueOrBitMapIndexList的基础上继续填充
			/*示例 三维位图 A-B-C
			条件为  or:{A in [A1]}
			则C 维度的[[1,0], [1,1], [1,2], [1,3]]都需要填充
			也就是
			[1,0,0], [1,1,0], [1,2,0], [1,3,0]
			[1,0,1], [1,1,1], [1,2,1], [1,3,1]
			[1,0,2], [1,1,2], [1,2,2], [1,3,2]
				 C	   C1               C2               C3
				A/B B0 B1 B2 B3  A/B B0 B1 B2 B3  A/B B0 B1 B2 B3
				A0  0  0  0  0   A0  0  0  0  0   A0  0  0  0  0
				A1  1  1  1  1   A1  1  1  1  1   A1  1  1  1  1
				A2  0  0  0  0   A2  0  0  0  0   A2  0  0  0  0
				A3  0  0  0  0   A3  0  0  0  0   A3  0  0  0  0
			*/
			tempIndexList := make([][]int64, 0)
			for i := 0; i < len(valueBitMapIndexList); i++ {
				for j := 0; j < 2*len(otherFuncDataScope)+1; j++ {
					copyList := make([]int64, len(functionIndexMap))
					copy(copyList, valueBitMapIndexList[i])
					copyList[functionIndexMap[function]] = functionValueIndexMap[function][value]
					copyList[functionIndexMap[otherFunction]] = int64(j)
					tempIndexList = append(tempIndexList, copyList)
				}
			}
			valueBitMapIndexList = tempIndexList
		}
		finalBitMapIndexList = append(finalBitMapIndexList, valueBitMapIndexList...)
	}
	return finalBitMapIndexList
}

func getGTBitMapIndexList(function string, rangeDataScope interface{},
	functionValueIndexMap map[string]map[interface{}]int64, rangeFunctionValueIndexMap map[string]map[interface{}]int64,
	functionIndexMap map[string]int64, equalFlag bool) [][]int64 {
	finalBitMapIndexList := make([][]int64, 0)
	// 遍历function data_scope 取值
	startValueIndex := rangeFunctionValueIndexMap[function][rangeDataScope]
	// 如果是gt, 则不包含当前取值下标
	if !equalFlag {
		startValueIndex += 1
	}
	for valueIndex := startValueIndex; valueIndex < int64(len(rangeFunctionValueIndexMap[function])*2+1); valueIndex++ {
		valueBitMapIndexList := make([][]int64, 0)
		for otherFunction, otherFuncDataScope := range functionValueIndexMap {
			if otherFunction == function {
				continue
			}
			//valueOrBitMapIndexList为空，进行初始化
			//or:{A in [A1]} valueBitMapIndexList = [[1,0], [1,1], [1,2], [1,3]]
			if len(valueBitMapIndexList) == 0 {
				for i := 0; i < len(otherFuncDataScope); i++ {
					valueBitMapIndexList = append(valueBitMapIndexList, make([]int64, len(functionIndexMap)))
					valueBitMapIndexList[i][functionIndexMap[function]] = valueIndex
					valueBitMapIndexList[i][functionIndexMap[otherFunction]] = int64(i)
				}
				continue
			}
			//若还有其他维度，则需要在valueOrBitMapIndexList的基础上继续填充
			/*示例 三维位图 A-B-C
			条件为  or:{A in [A1]}
			则C 维度的[[1,0], [1,1], [1,2], [1,3]]都需要填充
			也就是
			[1,0,0], [1,1,0], [1,2,0], [1,3,0]
			[1,0,1], [1,1,1], [1,2,1], [1,3,1]
			[1,0,2], [1,1,2], [1,2,2], [1,3,2]
				 C	   C1               C2               C3
				A/B B0 B1 B2 B3  A/B B0 B1 B2 B3  A/B B0 B1 B2 B3
				A0  0  0  0  0   A0  0  0  0  0   A0  0  0  0  0
				A1  1  1  1  1   A1  1  1  1  1   A1  1  1  1  1
				A2  0  0  0  0   A2  0  0  0  0   A2  0  0  0  0
				A3  0  0  0  0   A3  0  0  0  0   A3  0  0  0  0
			*/
			tempIndexList := make([][]int64, 0)
			for i := 0; i < len(valueBitMapIndexList); i++ {
				for j := 0; j < len(otherFuncDataScope); j++ {
					copyList := make([]int64, len(functionIndexMap))
					copy(copyList, valueBitMapIndexList[i])
					copyList[functionIndexMap[function]] = valueIndex
					copyList[functionIndexMap[otherFunction]] = int64(j)
					tempIndexList = append(tempIndexList, copyList)
				}
			}
			valueBitMapIndexList = tempIndexList
		}
		//处理范围值列表 rangeFunctionValueIndexMap
		for otherFunction, otherFuncDataScope := range rangeFunctionValueIndexMap {
			if otherFunction == function {
				continue
			}
			//valueOrBitMapIndexList为空，进行初始化
			//or:{A in [A1]} valueBitMapIndexList = [[1,0], [1,1], [1,2], [1,3]]
			if len(valueBitMapIndexList) == 0 {
				for i := 0; i < 2*len(otherFuncDataScope)+1; i++ {
					valueBitMapIndexList = append(valueBitMapIndexList, make([]int64, len(functionIndexMap)))
					valueBitMapIndexList[i][functionIndexMap[function]] = valueIndex
					valueBitMapIndexList[i][functionIndexMap[otherFunction]] = int64(i)
				}
				continue
			}
			//若还有其他维度，则需要在valueOrBitMapIndexList的基础上继续填充
			/*示例 三维位图 A-B-C
			条件为  or:{A in [A1]}
			则C 维度的[[1,0], [1,1], [1,2], [1,3]]都需要填充
			也就是
			[1,0,0], [1,1,0], [1,2,0], [1,3,0]
			[1,0,1], [1,1,1], [1,2,1], [1,3,1]
			[1,0,2], [1,1,2], [1,2,2], [1,3,2]
				 C	   C1               C2               C3
				A/B B0 B1 B2 B3  A/B B0 B1 B2 B3  A/B B0 B1 B2 B3
				A0  0  0  0  0   A0  0  0  0  0   A0  0  0  0  0
				A1  1  1  1  1   A1  1  1  1  1   A1  1  1  1  1
				A2  0  0  0  0   A2  0  0  0  0   A2  0  0  0  0
				A3  0  0  0  0   A3  0  0  0  0   A3  0  0  0  0
			*/
			tempIndexList := make([][]int64, 0)
			for i := 0; i < len(valueBitMapIndexList); i++ {
				for j := 0; j < 2*len(otherFuncDataScope)+1; j++ {
					copyList := make([]int64, len(functionIndexMap))
					copy(copyList, valueBitMapIndexList[i])
					copyList[functionIndexMap[function]] = valueIndex
					copyList[functionIndexMap[otherFunction]] = int64(j)
					tempIndexList = append(tempIndexList, copyList)
				}
			}
			valueBitMapIndexList = tempIndexList
		}
		finalBitMapIndexList = append(finalBitMapIndexList, valueBitMapIndexList...)
	}
	return finalBitMapIndexList
}

func getLTBitMapIndexList(function string, rangeDataScope interface{},
	functionValueIndexMap map[string]map[interface{}]int64, rangeFunctionValueIndexMap map[string]map[interface{}]int64,
	functionIndexMap map[string]int64, equalFlag bool) [][]int64 {
	finalBitMapIndexList := make([][]int64, 0)
	// 遍历function data_scope 取值
	startValueIndex := rangeFunctionValueIndexMap[function][rangeDataScope]
	// 如果是gt, 则不包含当前取值下标
	if !equalFlag {
		startValueIndex -= 1
	}
	for valueIndex := int64(0); valueIndex <= startValueIndex; valueIndex++ {
		valueBitMapIndexList := make([][]int64, 0)
		for otherFunction, otherFuncDataScope := range functionValueIndexMap {
			if otherFunction == function {
				continue
			}
			//valueOrBitMapIndexList为空，进行初始化
			//or:{A in [A1]} valueBitMapIndexList = [[1,0], [1,1], [1,2], [1,3]]
			if len(valueBitMapIndexList) == 0 {
				for i := 0; i < len(otherFuncDataScope); i++ {
					valueBitMapIndexList = append(valueBitMapIndexList, make([]int64, len(functionIndexMap)))
					valueBitMapIndexList[i][functionIndexMap[function]] = valueIndex
					valueBitMapIndexList[i][functionIndexMap[otherFunction]] = int64(i)
				}
				continue
			}
			//若还有其他维度，则需要在valueOrBitMapIndexList的基础上继续填充
			/*示例 三维位图 A-B-C
			条件为  or:{A in [A1]}
			则C 维度的[[1,0], [1,1], [1,2], [1,3]]都需要填充
			也就是
			[1,0,0], [1,1,0], [1,2,0], [1,3,0]
			[1,0,1], [1,1,1], [1,2,1], [1,3,1]
			[1,0,2], [1,1,2], [1,2,2], [1,3,2]
				 C	   C1               C2               C3
				A/B B0 B1 B2 B3  A/B B0 B1 B2 B3  A/B B0 B1 B2 B3
				A0  0  0  0  0   A0  0  0  0  0   A0  0  0  0  0
				A1  1  1  1  1   A1  1  1  1  1   A1  1  1  1  1
				A2  0  0  0  0   A2  0  0  0  0   A2  0  0  0  0
				A3  0  0  0  0   A3  0  0  0  0   A3  0  0  0  0
			*/
			tempIndexList := make([][]int64, 0)
			for i := 0; i < len(valueBitMapIndexList); i++ {
				for j := 0; j < len(otherFuncDataScope); j++ {
					copyList := make([]int64, len(functionIndexMap))
					copy(copyList, valueBitMapIndexList[i])
					copyList[functionIndexMap[function]] = valueIndex
					copyList[functionIndexMap[otherFunction]] = int64(j)
					tempIndexList = append(tempIndexList, copyList)
				}
			}
			valueBitMapIndexList = tempIndexList
		}
		//处理范围值列表 rangeFunctionValueIndexMap
		for otherFunction, otherFuncDataScope := range rangeFunctionValueIndexMap {
			if otherFunction == function {
				continue
			}
			//valueOrBitMapIndexList为空，进行初始化
			//or:{A in [A1]} valueBitMapIndexList = [[1,0], [1,1], [1,2], [1,3]]
			if len(valueBitMapIndexList) == 0 {
				for i := 0; i < 2*len(otherFuncDataScope)+1; i++ {
					valueBitMapIndexList = append(valueBitMapIndexList, make([]int64, len(functionIndexMap)))
					valueBitMapIndexList[i][functionIndexMap[function]] = valueIndex
					valueBitMapIndexList[i][functionIndexMap[otherFunction]] = int64(i)
				}
				continue
			}
			//若还有其他维度，则需要在valueOrBitMapIndexList的基础上继续填充
			/*示例 三维位图 A-B-C
			条件为  or:{A in [A1]}
			则C 维度的[[1,0], [1,1], [1,2], [1,3]]都需要填充
			也就是
			[1,0,0], [1,1,0], [1,2,0], [1,3,0]
			[1,0,1], [1,1,1], [1,2,1], [1,3,1]
			[1,0,2], [1,1,2], [1,2,2], [1,3,2]
				 C	   C1               C2               C3
				A/B B0 B1 B2 B3  A/B B0 B1 B2 B3  A/B B0 B1 B2 B3
				A0  0  0  0  0   A0  0  0  0  0   A0  0  0  0  0
				A1  1  1  1  1   A1  1  1  1  1   A1  1  1  1  1
				A2  0  0  0  0   A2  0  0  0  0   A2  0  0  0  0
				A3  0  0  0  0   A3  0  0  0  0   A3  0  0  0  0
			*/
			tempIndexList := make([][]int64, 0)
			for i := 0; i < len(valueBitMapIndexList); i++ {
				for j := 0; j < 2*len(otherFuncDataScope)+1; j++ {
					copyList := make([]int64, len(functionIndexMap))
					copy(copyList, valueBitMapIndexList[i])
					copyList[functionIndexMap[function]] = valueIndex
					copyList[functionIndexMap[otherFunction]] = int64(j)
					tempIndexList = append(tempIndexList, copyList)
				}
			}
			valueBitMapIndexList = tempIndexList
		}
		finalBitMapIndexList = append(finalBitMapIndexList, valueBitMapIndexList...)
	}
	return finalBitMapIndexList
}

func getBitMapLengthList(functionValueIndexMap map[string]map[interface{}]int64,
	rangeFunctionValueIndexMap map[string]map[interface{}]int64, functionIndexMap map[string]int64) []int64 {
	//根据function及各function data_scope长度 初始化多维位图
	lengthList := make([]int64, len(functionIndexMap))
	for function, dataScope := range functionValueIndexMap {
		lengthList[functionIndexMap[function]] = int64(len(dataScope))
	}
	for function, dataScope := range rangeFunctionValueIndexMap {
		lengthList[functionIndexMap[function]] = int64(2*len(dataScope) + 1)
	}
	return lengthList
}

func deepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = deepCopy(v)
		}
		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = deepCopy(v)
		}
		return newSlice
	}
	return value
}
