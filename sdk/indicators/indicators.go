package indicatorsSDK

import (
	"math"

	"github.com/jinzhu/copier"
	fyersTypes "github.com/pranavsindura/at-watch/types/fyers"
	indicatorTypes "github.com/pranavsindura/at-watch/types/indicators"
	"github.com/pranavsindura/at-watch/utils"
)

func GetSuperTrend(
	prevCandle *fyersTypes.FyersHistoricalCandle,
	prevSuperTrendData *indicatorTypes.SuperTrendData,
	currentCandle *fyersTypes.FyersHistoricalCandle,
	config indicatorTypes.SuperTrendConfig,
) *indicatorTypes.SuperTrendData {

	currentIndex := 0
	pastTRList := make([]float64, 0)
	prevUpperBound := 0.
	prevLowerBound := 0.
	prevSuperTrend := 0.
	prevClose := 0.
	isUsable := false
	if prevSuperTrendData != nil {
		currentIndex = prevSuperTrendData.Index + 1
		copier.Copy(&pastTRList, prevSuperTrendData.PastTRList)
		prevUpperBound = prevSuperTrendData.FinalUpperBound
		prevLowerBound = prevSuperTrendData.FinalLowerBound
		prevSuperTrend = prevSuperTrendData.SuperTrend
	}
	if prevCandle != nil {
		prevClose = prevCandle.Close
	}

	tr := math.Max(math.Max(currentCandle.High-currentCandle.Low, math.Abs(currentCandle.High-prevClose)), math.Abs(currentCandle.Low-prevClose))

	if currentIndex < config.Period {
		pastTRList = append(pastTRList, tr)
	} else if currentIndex >= config.Period {
		pastTRList = pastTRList[1:]
		pastTRList = append(pastTRList, tr)
		isUsable = true
	}

	atr := 0.

	for _, pastTR := range pastTRList {
		atr += pastTR / float64(config.Period)
	}

	basicUpperBound := ((currentCandle.High + currentCandle.Low) / 2.) + atr*config.Multiplier
	basicLowerBound := ((currentCandle.High + currentCandle.Low) / 2.) - atr*config.Multiplier

	finalUpperBound := 0.
	finalLowerBound := 0.
	superTrend := 0.
	superTrendDirection := false

	if basicUpperBound < prevUpperBound || (prevCandle != nil && prevCandle.Close > prevUpperBound) {
		finalUpperBound = basicUpperBound
	} else {
		finalUpperBound = prevUpperBound
	}

	if basicLowerBound > prevLowerBound || (prevCandle != nil && prevCandle.Close < prevLowerBound) {
		finalLowerBound = basicLowerBound
	} else {
		finalLowerBound = prevLowerBound
	}

	if utils.FloatCompare(prevSuperTrend, prevUpperBound) == 0 && currentCandle.Close < finalUpperBound {
		superTrend = finalUpperBound
		superTrendDirection = false
	} else if utils.FloatCompare(prevSuperTrend, prevUpperBound) == 0 && currentCandle.Close >= finalUpperBound {
		superTrend = finalLowerBound
		superTrendDirection = true
	} else if utils.FloatCompare(prevSuperTrend, prevLowerBound) == 0 && currentCandle.Close > finalLowerBound {
		superTrend = finalLowerBound
		superTrendDirection = true
	} else if utils.FloatCompare(prevSuperTrend, prevLowerBound) == 0 && currentCandle.Close <= finalLowerBound {
		superTrend = finalUpperBound
		superTrendDirection = false
	}

	superTrendData := &indicatorTypes.SuperTrendData{
		Index:               currentIndex,
		ATR:                 atr,
		PastTRList:          pastTRList,
		BasicUpperBound:     basicUpperBound,
		BasicLowerBound:     basicLowerBound,
		FinalUpperBound:     finalUpperBound,
		FinalLowerBound:     finalLowerBound,
		SuperTrend:          superTrend,
		SuperTrendDirection: superTrendDirection,
		IsUsable:            isUsable,
	}

	return superTrendData
}
