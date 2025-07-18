package persistence

import (
	"context"
	"encoding/json"

	"github.com/Masterminds/squirrel"
	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/conf/configtest"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/model/request"
	"github.com/navidrome/navidrome/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ArtistRepository", func() {
	var repo model.ArtistRepository

	BeforeEach(func() {
		DeferCleanup(configtest.SetupConfig())
		ctx := log.NewContext(context.TODO())
		ctx = request.WithUser(ctx, model.User{ID: "userid"})
		repo = NewArtistRepository(ctx, GetDBXBuilder())
	})

	Describe("Count", func() {
		It("returns the number of artists in the DB", func() {
			Expect(repo.CountAll()).To(Equal(int64(2)))
		})
	})

	Describe("Exists", func() {
		It("returns true for an artist that is in the DB", func() {
			Expect(repo.Exists("3")).To(BeTrue())
		})
		It("returns false for an artist that is in the DB", func() {
			Expect(repo.Exists("666")).To(BeFalse())
		})
	})

	Describe("Get", func() {
		It("saves and retrieves data", func() {
			artist, err := repo.Get("2")
			Expect(err).ToNot(HaveOccurred())
			Expect(artist.Name).To(Equal(artistKraftwerk.Name))
		})
	})

	Describe("GetIndexKey", func() {
		// Note: OrderArtistName should never be empty, so we don't need to test for that
		r := artistRepository{indexGroups: utils.ParseIndexGroups(conf.Server.IndexGroups)}
		When("PreferSortTags is false", func() {
			BeforeEach(func() {
				conf.Server.PreferSortTags = false
			})
			It("returns the OrderArtistName key is SortArtistName is empty", func() {
				conf.Server.PreferSortTags = false
				a := model.Artist{SortArtistName: "", OrderArtistName: "Bar", Name: "Qux"}
				idx := GetIndexKey(&r, a)
				Expect(idx).To(Equal("B"))
			})
			It("returns the OrderArtistName key even if SortArtistName is not empty", func() {
				a := model.Artist{SortArtistName: "Foo", OrderArtistName: "Bar", Name: "Qux"}
				idx := GetIndexKey(&r, a)
				Expect(idx).To(Equal("B"))
			})
		})
		When("PreferSortTags is true", func() {
			BeforeEach(func() {
				conf.Server.PreferSortTags = true
			})
			It("returns the SortArtistName key if it is not empty", func() {
				a := model.Artist{SortArtistName: "Foo", OrderArtistName: "Bar", Name: "Qux"}
				idx := GetIndexKey(&r, a)
				Expect(idx).To(Equal("F"))
			})
			It("returns the OrderArtistName key if SortArtistName is empty", func() {
				a := model.Artist{SortArtistName: "", OrderArtistName: "Bar", Name: "Qux"}
				idx := GetIndexKey(&r, a)
				Expect(idx).To(Equal("B"))
			})
		})
	})

	Describe("GetIndex", func() {
		When("PreferSortTags is true", func() {
			BeforeEach(func() {
				conf.Server.PreferSortTags = true
			})
			It("returns the index when PreferSortTags is true and SortArtistName is not empty", func() {
				// Set SortArtistName to "Foo" for Beatles
				artistBeatles.SortArtistName = "Foo"
				er := repo.Put(&artistBeatles)
				Expect(er).To(BeNil())

				idx, err := repo.GetIndex(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(2))
				Expect(idx[0].ID).To(Equal("F"))
				Expect(idx[0].Artists).To(HaveLen(1))
				Expect(idx[0].Artists[0].Name).To(Equal(artistBeatles.Name))
				Expect(idx[1].ID).To(Equal("K"))
				Expect(idx[1].Artists).To(HaveLen(1))
				Expect(idx[1].Artists[0].Name).To(Equal(artistKraftwerk.Name))

				// Restore the original value
				artistBeatles.SortArtistName = ""
				er = repo.Put(&artistBeatles)
				Expect(er).To(BeNil())
			})

			// BFR Empty SortArtistName is not saved in the DB anymore
			XIt("returns the index when PreferSortTags is true and SortArtistName is empty", func() {
				idx, err := repo.GetIndex(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(2))
				Expect(idx[0].ID).To(Equal("B"))
				Expect(idx[0].Artists).To(HaveLen(1))
				Expect(idx[0].Artists[0].Name).To(Equal(artistBeatles.Name))
				Expect(idx[1].ID).To(Equal("K"))
				Expect(idx[1].Artists).To(HaveLen(1))
				Expect(idx[1].Artists[0].Name).To(Equal(artistKraftwerk.Name))
			})
		})

		When("PreferSortTags is false", func() {
			BeforeEach(func() {
				conf.Server.PreferSortTags = false
			})
			It("returns the index when SortArtistName is NOT empty", func() {
				// Set SortArtistName to "Foo" for Beatles
				artistBeatles.SortArtistName = "Foo"
				er := repo.Put(&artistBeatles)
				Expect(er).To(BeNil())

				idx, err := repo.GetIndex(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(2))
				Expect(idx[0].ID).To(Equal("B"))
				Expect(idx[0].Artists).To(HaveLen(1))
				Expect(idx[0].Artists[0].Name).To(Equal(artistBeatles.Name))
				Expect(idx[1].ID).To(Equal("K"))
				Expect(idx[1].Artists).To(HaveLen(1))
				Expect(idx[1].Artists[0].Name).To(Equal(artistKraftwerk.Name))

				// Restore the original value
				artistBeatles.SortArtistName = ""
				er = repo.Put(&artistBeatles)
				Expect(er).To(BeNil())
			})

			It("returns the index when SortArtistName is empty", func() {
				idx, err := repo.GetIndex(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(2))
				Expect(idx[0].ID).To(Equal("B"))
				Expect(idx[0].Artists).To(HaveLen(1))
				Expect(idx[0].Artists[0].Name).To(Equal(artistBeatles.Name))
				Expect(idx[1].ID).To(Equal("K"))
				Expect(idx[1].Artists).To(HaveLen(1))
				Expect(idx[1].Artists[0].Name).To(Equal(artistKraftwerk.Name))
			})
		})

		When("filtering by role", func() {
			var raw *artistRepository

			BeforeEach(func() {
				raw = repo.(*artistRepository)
				// Add stats to artists using direct SQL since Put doesn't populate stats
				composerStats := `{"composer": {"s": 1000, "m": 5, "a": 2}}`
				producerStats := `{"producer": {"s": 500, "m": 3, "a": 1}}`

				// Set Beatles as composer
				_, err := raw.executeSQL(squirrel.Update(raw.tableName).Set("stats", composerStats).Where(squirrel.Eq{"id": artistBeatles.ID}))
				Expect(err).ToNot(HaveOccurred())

				// Set Kraftwerk as producer
				_, err = raw.executeSQL(squirrel.Update(raw.tableName).Set("stats", producerStats).Where(squirrel.Eq{"id": artistKraftwerk.ID}))
				Expect(err).ToNot(HaveOccurred())
			})

			AfterEach(func() {
				// Clean up stats
				_, _ = raw.executeSQL(squirrel.Update(raw.tableName).Set("stats", "{}").Where(squirrel.Eq{"id": artistBeatles.ID}))
				_, _ = raw.executeSQL(squirrel.Update(raw.tableName).Set("stats", "{}").Where(squirrel.Eq{"id": artistKraftwerk.ID}))
			})

			It("returns only artists with the specified role", func() {
				idx, err := repo.GetIndex(false, model.RoleComposer)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(1))
				Expect(idx[0].ID).To(Equal("B"))
				Expect(idx[0].Artists).To(HaveLen(1))
				Expect(idx[0].Artists[0].Name).To(Equal(artistBeatles.Name))
			})

			It("returns artists with any of the specified roles", func() {
				idx, err := repo.GetIndex(false, model.RoleComposer, model.RoleProducer)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(2))

				// Find Beatles and Kraftwerk in the results
				var beatlesFound, kraftwerkFound bool
				for _, index := range idx {
					for _, artist := range index.Artists {
						if artist.Name == artistBeatles.Name {
							beatlesFound = true
						}
						if artist.Name == artistKraftwerk.Name {
							kraftwerkFound = true
						}
					}
				}
				Expect(beatlesFound).To(BeTrue())
				Expect(kraftwerkFound).To(BeTrue())
			})

			It("returns empty index when no artists have the specified role", func() {
				idx, err := repo.GetIndex(false, model.RoleDirector)
				Expect(err).ToNot(HaveOccurred())
				Expect(idx).To(HaveLen(0))
			})
		})
	})

	Describe("dbArtist mapping", func() {
		var (
			artist *model.Artist
			dba    *dbArtist
		)

		BeforeEach(func() {
			artist = &model.Artist{ID: "1", Name: "Eddie Van Halen", SortArtistName: "Van Halen, Eddie"}
			dba = &dbArtist{Artist: artist}
		})

		Describe("PostScan", func() {
			It("parses stats and similar artists correctly", func() {
				stats := map[string]map[string]int64{
					"total":    {"s": 1000, "m": 10, "a": 2},
					"composer": {"s": 500, "m": 5, "a": 1},
				}
				statsJSON, _ := json.Marshal(stats)
				dba.Stats = string(statsJSON)
				dba.SimilarArtists = `[{"id":"2","Name":"AC/DC"},{"name":"Test;With:Sep,Chars"}]`

				err := dba.PostScan()
				Expect(err).ToNot(HaveOccurred())
				Expect(dba.Artist.Size).To(Equal(int64(1000)))
				Expect(dba.Artist.SongCount).To(Equal(10))
				Expect(dba.Artist.AlbumCount).To(Equal(2))
				Expect(dba.Artist.Stats).To(HaveLen(1))
				Expect(dba.Artist.Stats[model.RoleFromString("composer")].Size).To(Equal(int64(500)))
				Expect(dba.Artist.Stats[model.RoleFromString("composer")].SongCount).To(Equal(5))
				Expect(dba.Artist.Stats[model.RoleFromString("composer")].AlbumCount).To(Equal(1))
				Expect(dba.Artist.SimilarArtists).To(HaveLen(2))
				Expect(dba.Artist.SimilarArtists[0].ID).To(Equal("2"))
				Expect(dba.Artist.SimilarArtists[0].Name).To(Equal("AC/DC"))
				Expect(dba.Artist.SimilarArtists[1].ID).To(BeEmpty())
				Expect(dba.Artist.SimilarArtists[1].Name).To(Equal("Test;With:Sep,Chars"))
			})
		})

		Describe("PostMapArgs", func() {
			It("maps empty similar artists correctly", func() {
				m := make(map[string]any)
				err := dba.PostMapArgs(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m).To(HaveKeyWithValue("similar_artists", "[]"))
			})

			It("maps similar artists and full text correctly", func() {
				artist.SimilarArtists = []model.Artist{
					{ID: "2", Name: "AC/DC"},
					{Name: "Test;With:Sep,Chars"},
				}
				m := make(map[string]any)
				err := dba.PostMapArgs(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m).To(HaveKeyWithValue("similar_artists", `[{"id":"2","name":"AC/DC"},{"name":"Test;With:Sep,Chars"}]`))
				Expect(m).To(HaveKeyWithValue("full_text", " eddie halen van"))
			})

			It("does not override empty sort_artist_name and mbz_artist_id", func() {
				m := map[string]any{
					"sort_artist_name": "",
					"mbz_artist_id":    "",
				}
				err := dba.PostMapArgs(m)
				Expect(err).ToNot(HaveOccurred())
				Expect(m).ToNot(HaveKey("sort_artist_name"))
				Expect(m).ToNot(HaveKey("mbz_artist_id"))
			})
		})

		Describe("Missing artist visibility", func() {
			var raw *artistRepository
			var missing model.Artist

			insertMissing := func() {
				missing = model.Artist{ID: "m1", Name: "Missing", OrderArtistName: "missing"}
				Expect(repo.Put(&missing)).To(Succeed())
				raw = repo.(*artistRepository)
				_, err := raw.executeSQL(squirrel.Update(raw.tableName).Set("missing", true).Where(squirrel.Eq{"id": missing.ID}))
				Expect(err).ToNot(HaveOccurred())
			}

			removeMissing := func() {
				if raw != nil {
					_, _ = raw.executeSQL(squirrel.Delete(raw.tableName).Where(squirrel.Eq{"id": missing.ID}))
				}
			}

			Context("regular user", func() {
				BeforeEach(func() {
					ctx := log.NewContext(context.TODO())
					ctx = request.WithUser(ctx, model.User{ID: "u1"})
					repo = NewArtistRepository(ctx, GetDBXBuilder())
					insertMissing()
				})

				AfterEach(func() { removeMissing() })

				It("does not return missing artist in GetAll", func() {
					artists, err := repo.GetAll(model.QueryOptions{Filters: squirrel.Eq{"artist.missing": false}})
					Expect(err).ToNot(HaveOccurred())
					Expect(artists).To(HaveLen(2))
				})

				It("does not return missing artist in Search", func() {
					res, err := repo.Search("missing", 0, 10, false)
					Expect(err).ToNot(HaveOccurred())
					Expect(res).To(BeEmpty())
				})

				It("does not return missing artist in GetIndex", func() {
					idx, err := repo.GetIndex(false)
					Expect(err).ToNot(HaveOccurred())
					// Only 2 artists should be present
					total := 0
					for _, ix := range idx {
						total += len(ix.Artists)
					}
					Expect(total).To(Equal(2))
				})
			})

			Context("admin user", func() {
				BeforeEach(func() {
					ctx := log.NewContext(context.TODO())
					ctx = request.WithUser(ctx, model.User{ID: "admin", IsAdmin: true})
					repo = NewArtistRepository(ctx, GetDBXBuilder())
					insertMissing()
				})

				AfterEach(func() { removeMissing() })

				It("returns missing artist in GetAll", func() {
					artists, err := repo.GetAll()
					Expect(err).ToNot(HaveOccurred())
					Expect(artists).To(HaveLen(3))
				})

				It("returns missing artist in Search", func() {
					res, err := repo.Search("missing", 0, 10, true)
					Expect(err).ToNot(HaveOccurred())
					Expect(res).To(HaveLen(1))
				})

				It("returns missing artist in GetIndex when included", func() {
					idx, err := repo.GetIndex(true)
					Expect(err).ToNot(HaveOccurred())
					total := 0
					for _, ix := range idx {
						total += len(ix.Artists)
					}
					Expect(total).To(Equal(3))
				})
			})
		})
	})

	Describe("roleFilter", func() {
		It("filters out roles not present in the participants model", func() {
			Expect(roleFilter("", "artist")).To(Equal(squirrel.NotEq{"stats ->> '$.artist'": nil}))
			Expect(roleFilter("", "albumartist")).To(Equal(squirrel.NotEq{"stats ->> '$.albumartist'": nil}))
			Expect(roleFilter("", "composer")).To(Equal(squirrel.NotEq{"stats ->> '$.composer'": nil}))
			Expect(roleFilter("", "conductor")).To(Equal(squirrel.NotEq{"stats ->> '$.conductor'": nil}))
			Expect(roleFilter("", "lyricist")).To(Equal(squirrel.NotEq{"stats ->> '$.lyricist'": nil}))
			Expect(roleFilter("", "arranger")).To(Equal(squirrel.NotEq{"stats ->> '$.arranger'": nil}))
			Expect(roleFilter("", "producer")).To(Equal(squirrel.NotEq{"stats ->> '$.producer'": nil}))
			Expect(roleFilter("", "director")).To(Equal(squirrel.NotEq{"stats ->> '$.director'": nil}))
			Expect(roleFilter("", "engineer")).To(Equal(squirrel.NotEq{"stats ->> '$.engineer'": nil}))
			Expect(roleFilter("", "mixer")).To(Equal(squirrel.NotEq{"stats ->> '$.mixer'": nil}))
			Expect(roleFilter("", "remixer")).To(Equal(squirrel.NotEq{"stats ->> '$.remixer'": nil}))
			Expect(roleFilter("", "djmixer")).To(Equal(squirrel.NotEq{"stats ->> '$.djmixer'": nil}))
			Expect(roleFilter("", "performer")).To(Equal(squirrel.NotEq{"stats ->> '$.performer'": nil}))

			Expect(roleFilter("", "wizard")).To(Equal(squirrel.Eq{"1": 2}))
			Expect(roleFilter("", "songanddanceman")).To(Equal(squirrel.Eq{"1": 2}))
			Expect(roleFilter("", "artist') SELECT LIKE(CHAR(65,66,67,68,69,70,71),UPPER(HEX(RANDOMBLOB(500000000/2))))--")).To(Equal(squirrel.Eq{"1": 2}))
		})
	})

	Context("MBID Search", func() {
		var artistWithMBID model.Artist
		var raw *artistRepository

		BeforeEach(func() {
			raw = repo.(*artistRepository)
			// Create a test artist with MBID
			artistWithMBID = model.Artist{
				ID:          "test-mbid-artist",
				Name:        "Test MBID Artist",
				MbzArtistID: "550e8400-e29b-41d4-a716-446655440010", // Valid UUID v4
			}

			// Insert the test artist into the database
			err := repo.Put(&artistWithMBID)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			// Clean up test data using direct SQL
			_, _ = raw.executeSQL(squirrel.Delete(raw.tableName).Where(squirrel.Eq{"id": artistWithMBID.ID}))
		})

		It("finds artist by mbz_artist_id", func() {
			results, err := repo.Search("550e8400-e29b-41d4-a716-446655440010", 0, 10, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0].ID).To(Equal("test-mbid-artist"))
			Expect(results[0].Name).To(Equal("Test MBID Artist"))
		})

		It("returns empty result when MBID is not found", func() {
			results, err := repo.Search("550e8400-e29b-41d4-a716-446655440099", 0, 10, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(BeEmpty())
		})

		It("handles includeMissing parameter for MBID search", func() {
			// Create a missing artist with MBID
			missingArtist := model.Artist{
				ID:          "test-missing-mbid-artist",
				Name:        "Test Missing MBID Artist",
				MbzArtistID: "550e8400-e29b-41d4-a716-446655440012",
				Missing:     true,
			}

			err := repo.Put(&missingArtist)
			Expect(err).ToNot(HaveOccurred())

			// Should not find missing artist when includeMissing is false
			results, err := repo.Search("550e8400-e29b-41d4-a716-446655440012", 0, 10, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(BeEmpty())

			// Should find missing artist when includeMissing is true
			results, err = repo.Search("550e8400-e29b-41d4-a716-446655440012", 0, 10, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(1))
			Expect(results[0].ID).To(Equal("test-missing-mbid-artist"))

			// Clean up
			_, _ = raw.executeSQL(squirrel.Delete(raw.tableName).Where(squirrel.Eq{"id": missingArtist.ID}))
		})
	})
})
