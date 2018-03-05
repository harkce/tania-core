package sqlite

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/Tanibox/tania-server/src/assets/domain"
	"github.com/Tanibox/tania-server/src/assets/query"
	"github.com/Tanibox/tania-server/src/assets/storage"
	"github.com/Tanibox/tania-server/src/helper/paginationhelper"
	uuid "github.com/satori/go.uuid"
)

type MaterialReadQuerySqlite struct {
	DB *sql.DB
}

func NewMaterialReadQuerySqlite(db *sql.DB) query.MaterialReadQuery {
	return MaterialReadQuerySqlite{DB: db}
}

type materialReadResult struct {
	UID            string
	Name           string
	PricePerUnit   string
	CurrencyCode   string
	Type           string
	TypeData       string
	Quantity       float32
	QuantityUnit   string
	ExpirationDate sql.NullString
	Notes          sql.NullString
	ProducedBy     sql.NullString
	CreatedDate    string
}

func (q MaterialReadQuerySqlite) FindAll(materialType, materialTypeDetail string, page, limit int) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		materialReads := []storage.MaterialRead{}
		rowsData := materialReadResult{}
		var params []interface{}

		sql := "SELECT * FROM MATERIAL_READ WHERE 1 = 1"

		if materialType != "" {
			t := strings.Split(materialType, ",")

			sql += " AND TYPE = ?"
			params = append(params, t[0])

			for _, v := range t[1:] {
				sql += " OR TYPE = ?"
				params = append(params, v)
			}
		}
		if materialTypeDetail != "" {
			t := strings.Split(materialTypeDetail, ",")

			sql += " AND TYPE_DATA = ?"
			params = append(params, t[0])

			for _, v := range t[1:] {
				sql += " OR TYPE_DATA = ?"
				params = append(params, v)
			}
		}

		offset := paginationhelper.CalculatePageToOffset(page, limit)

		sql += " ORDER BY CREATED_DATE DESC LIMIT ? OFFSET ?"
		params = append(params, limit, offset)

		rows, err := q.DB.Query(sql, params...)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		for rows.Next() {
			err = rows.Scan(
				&rowsData.UID,
				&rowsData.Name,
				&rowsData.PricePerUnit,
				&rowsData.CurrencyCode,
				&rowsData.Type,
				&rowsData.TypeData,
				&rowsData.Quantity,
				&rowsData.QuantityUnit,
				&rowsData.ExpirationDate,
				&rowsData.Notes,
				&rowsData.ProducedBy,
				&rowsData.CreatedDate,
			)

			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			materialUID, err := uuid.FromString(rowsData.UID)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			var mExpDate *time.Time
			if rowsData.ExpirationDate.Valid && rowsData.ExpirationDate.String != "" {
				date, err := time.Parse(time.RFC3339, rowsData.ExpirationDate.String)
				if err != nil {
					result <- query.QueryResult{Error: err}
				}

				mExpDate = &date
			}

			mCreatedDate, err := time.Parse(time.RFC3339, rowsData.CreatedDate)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			pricePerUnit, err := domain.CreatePricePerUnit(rowsData.PricePerUnit, rowsData.CurrencyCode)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			var materialType storage.MaterialType
			switch rowsData.Type {
			case domain.MaterialTypePlantCode:
				materialType, err = domain.CreateMaterialTypePlant(rowsData.TypeData)
				if err != nil {
					result <- query.QueryResult{Error: err}
				}
			case domain.MaterialTypeSeedCode:
				materialType, err = domain.CreateMaterialTypeSeed(rowsData.TypeData)
				if err != nil {
					result <- query.QueryResult{Error: err}
				}
			case domain.MaterialTypeGrowingMediumCode:
				materialType = domain.MaterialTypeGrowingMedium{}
			case domain.MaterialTypeAgrochemicalCode:
				materialType, err = domain.CreateMaterialTypeAgrochemical(rowsData.TypeData)
				if err != nil {
					result <- query.QueryResult{Error: err}
				}
			case domain.MaterialTypeLabelAndCropSupportCode:
				materialType = domain.MaterialTypeLabelAndCropSupport{}
			case domain.MaterialTypeSeedingContainerCode:
				materialType, err = domain.CreateMaterialTypeSeedingContainer(rowsData.TypeData)
				if err != nil {
					result <- query.QueryResult{Error: err}
				}
			case domain.MaterialTypePostHarvestSupplyCode:
				materialType = domain.MaterialTypePostHarvestSupply{}
			case domain.MaterialTypeOtherCode:
				materialType = domain.MaterialTypeOther{}
			default:
				result <- query.QueryResult{Error: errors.New("Invalid material type")}
			}

			qtyUnit := domain.GetMaterialQuantityUnit(rowsData.Type, rowsData.QuantityUnit)
			if qtyUnit == (domain.MaterialQuantityUnit{}) {
				result <- query.QueryResult{Error: errors.New("Invalid quantity unit")}
			}

			var notes *string
			if rowsData.Notes.Valid {
				notes = &rowsData.Notes.String
			}

			var producedBy *string
			if rowsData.ProducedBy.Valid {
				producedBy = &rowsData.ProducedBy.String
			}

			materialReads = append(materialReads, storage.MaterialRead{
				UID:          materialUID,
				Name:         rowsData.Name,
				PricePerUnit: storage.PricePerUnit(pricePerUnit),
				Type:         materialType,
				Quantity: storage.MaterialQuantity{
					Unit:  qtyUnit,
					Value: rowsData.Quantity,
				},
				ExpirationDate: mExpDate,
				Notes:          notes,
				ProducedBy:     producedBy,
				CreatedDate:    mCreatedDate,
			})
		}

		result <- query.QueryResult{Result: materialReads}
		close(result)
	}()

	return result
}

func (q MaterialReadQuerySqlite) CountAll(materialType, materialTypeDetail string) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		total := 0
		var params []interface{}

		sql := "SELECT COUNT(UID) FROM MATERIAL_READ WHERE 1 = 1"

		if materialType != "" {
			t := strings.Split(materialType, ",")

			sql += " AND TYPE = ?"
			params = append(params, t[0])

			for _, v := range t[1:] {
				sql += " OR TYPE = ?"
				params = append(params, v)
			}
		}
		if materialTypeDetail != "" {
			t := strings.Split(materialTypeDetail, ",")

			sql += " AND TYPE_DATA = ?"
			params = append(params, t[0])

			for _, v := range t[1:] {
				sql += " OR TYPE_DATA = ?"
				params = append(params, v)
			}
		}

		err := q.DB.QueryRow(sql, params...).Scan(&total)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		result <- query.QueryResult{Result: total}
		close(result)
	}()

	return result
}

func (q MaterialReadQuerySqlite) FindByID(materialUID uuid.UUID) <-chan query.QueryResult {
	result := make(chan query.QueryResult)

	go func() {
		materialRead := storage.MaterialRead{}
		rowsData := materialReadResult{}

		err := q.DB.QueryRow(`SELECT * FROM MATERIAL_READ WHERE UID = ?`, materialUID).Scan(
			&rowsData.UID,
			&rowsData.Name,
			&rowsData.PricePerUnit,
			&rowsData.CurrencyCode,
			&rowsData.Type,
			&rowsData.TypeData,
			&rowsData.Quantity,
			&rowsData.QuantityUnit,
			&rowsData.ExpirationDate,
			&rowsData.Notes,
			&rowsData.ProducedBy,
			&rowsData.CreatedDate,
		)

		if err != nil && err != sql.ErrNoRows {
			result <- query.QueryResult{Error: err}
		}

		if err == sql.ErrNoRows {
			result <- query.QueryResult{Result: materialRead}
		}

		materialUID, err := uuid.FromString(rowsData.UID)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		var mExpDate *time.Time
		if rowsData.ExpirationDate.Valid && rowsData.ExpirationDate.String != "" {
			date, err := time.Parse(time.RFC3339, rowsData.ExpirationDate.String)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}

			mExpDate = &date
		}

		mCreatedDate, err := time.Parse(time.RFC3339, rowsData.CreatedDate)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		pricePerUnit, err := domain.CreatePricePerUnit(rowsData.PricePerUnit, rowsData.CurrencyCode)
		if err != nil {
			result <- query.QueryResult{Error: err}
		}

		var materialType storage.MaterialType
		switch rowsData.Type {
		case domain.MaterialTypePlantCode:
			materialType, err = domain.CreateMaterialTypePlant(rowsData.TypeData)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}
		case domain.MaterialTypeSeedCode:
			materialType, err = domain.CreateMaterialTypeSeed(rowsData.TypeData)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}
		case domain.MaterialTypeGrowingMediumCode:
			materialType = domain.MaterialTypeGrowingMedium{}
		case domain.MaterialTypeAgrochemicalCode:
			materialType, err = domain.CreateMaterialTypeAgrochemical(rowsData.TypeData)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}
		case domain.MaterialTypeLabelAndCropSupportCode:
			materialType = domain.MaterialTypeLabelAndCropSupport{}
		case domain.MaterialTypeSeedingContainerCode:
			materialType, err = domain.CreateMaterialTypeSeedingContainer(rowsData.TypeData)
			if err != nil {
				result <- query.QueryResult{Error: err}
			}
		case domain.MaterialTypePostHarvestSupplyCode:
			materialType = domain.MaterialTypePostHarvestSupply{}
		case domain.MaterialTypeOtherCode:
			materialType = domain.MaterialTypeOther{}
		default:
			result <- query.QueryResult{Error: errors.New("Invalid material type")}
		}

		qtyUnit := domain.GetMaterialQuantityUnit(rowsData.Type, rowsData.QuantityUnit)
		if qtyUnit == (domain.MaterialQuantityUnit{}) {
			result <- query.QueryResult{Error: errors.New("Invalid quantity unit")}
		}

		var notes *string
		if rowsData.Notes.Valid {
			notes = &rowsData.Notes.String
		}

		var producedBy *string
		if rowsData.ProducedBy.Valid {
			producedBy = &rowsData.ProducedBy.String
		}

		materialRead = storage.MaterialRead{
			UID:          materialUID,
			Name:         rowsData.Name,
			PricePerUnit: storage.PricePerUnit(pricePerUnit),
			Type:         materialType,
			Quantity: storage.MaterialQuantity{
				Unit:  qtyUnit,
				Value: rowsData.Quantity,
			},
			ExpirationDate: mExpDate,
			Notes:          notes,
			ProducedBy:     producedBy,
			CreatedDate:    mCreatedDate,
		}

		result <- query.QueryResult{Result: materialRead}
		close(result)
	}()

	return result
}
