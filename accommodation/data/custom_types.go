package data

import (
	"strconv"
	"strings"

	"github.com/gocql/gocql"
)

type CustomSetInt struct {
    Values []int
}

func (csi CustomSetInt) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
    if csi.Values == nil || len(csi.Values) == 0 {
        return nil, nil
    }

    // Konvertujemo sve vrednosti u string i spajamo ih zarezima.
    values := make([]string, len(csi.Values))
    for i, v := range csi.Values {
        values[i] = strconv.Itoa(v)
    }
    serialized := strings.Join(values, ",")

    return []byte(serialized), nil
}

func (csi *CustomSetInt) UnmarshalCQL(info gocql.TypeInfo, data []byte) error {
    if data == nil {
        csi.Values = nil
        return nil
    }

    // Razdvajamo string koristeći zareze.
    serialized := string(data)
    values := strings.Split(serialized, ",")

    // Konvertujemo string vrednosti u int.
    csi.Values = make([]int, len(values))
    for i, v := range values {
        intValue, err := strconv.Atoi(v)
        if err != nil {
            return err
        }
        csi.Values[i] = intValue
    }

    return nil
}

