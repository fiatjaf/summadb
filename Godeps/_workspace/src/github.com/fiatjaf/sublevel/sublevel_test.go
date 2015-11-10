package sublevel

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCRUD(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CRUD Suite")
}

var _ = Describe("CRUD", func() {
	var db AbstractLevel

	BeforeEach(func() {
		//os.RemoveAll("/tmp/subleveltesting.leveldb")
		db = OpenFile("/tmp/subleveltesting.leveldb", nil)
		Expect(db.err).To(BeNil())
	})

	AfterEach(func() {
		Expect(db.Close()).To(Succeed())
	})

	//Context("basic operations", func() {
	//	It("", func() {

	//	})
	//})

	//Context("list operations", func() {
	//	It("", func() {

	//	})
	//})

	Context("batch operations", func() {
		It("should execute operations on a single sublevel", func() {
			sub := db.MustSub("planoreal")
			b := sub.NewBatch()
			b.Put([]byte("persioaryda"), []byte("6"))
			b.Put([]byte("gustavofranco"), []byte("7"))
			b.Put([]byte("edmarbacha"), []byte("4"))
			b.Put([]byte("eugeniogudin"), []byte("8"))
			Expect(b.Len()).To(Equal(4))
			Expect(sub.Write(b, nil)).To(Succeed())
			Expect(sub.Get([]byte("eugeniogudin"), nil)).To(BeEquivalentTo("8"))
			Expect(sub.Get([]byte("persioaryda"), nil)).To(BeEquivalentTo("6"))

			c := sub.NewBatch()
			c.Put([]byte("andrelararesende"), []byte("3"))
			c.Delete([]byte("eugeniogudin"))
			Expect(c.Len()).To(Equal(2))
			Expect(sub.Write(c, nil)).To(Succeed())
			_, err := sub.Get([]byte("eugeniogudin"), nil)
			Expect(err).To(HaveOccurred())
			Expect(sub.Get([]byte("andrelararesende"), nil)).To(BeEquivalentTo("3"))
		})

		It("should execute operations on different sublevels", func() {
			sub1 := db.MustSub("planoreal")
			b1 := sub1.NewBatch()
			b1.Delete([]byte("persioaryda"))
			b1.Put([]byte("persioarida"), []byte("6"))
			b1.Put([]byte("pedromalan"), []byte("5"))

			sub2 := db.MustSub("planocruzado")
			b2 := sub2.NewBatch()
			b2.Put([]byte("persioarida"), []byte("2"))
			b2.Put([]byte("andrelararesende"), []byte("1"))

			b := db.NewBatch()
			b.MergeSubBatch(b1)
			b.MergeSubBatch(b2)
			Expect(db.Write(b, nil)).To(Succeed())

			Expect(sub1.Get([]byte("edmarbacha"), nil)).To(BeEquivalentTo("4"))
			Expect(sub1.Get([]byte("persioarida"), nil)).To(BeEquivalentTo("6"))
			Expect(sub1.Get([]byte("pedromalan"), nil)).To(BeEquivalentTo("5"))
			Expect(sub2.Get([]byte("persioarida"), nil)).To(BeEquivalentTo("2"))
			Expect(sub2.Get([]byte("andrelararesende"), nil)).To(BeEquivalentTo("1"))
		})

		It("should use different syntax for multisublevel batch", func() {
			sub1 := db.MustSub("planoreal")
			b1 := sub1.NewBatch()

			sub2 := db.MustSub("planocruzado")
			b2 := sub2.NewBatch()
			b2.Put([]byte("joaosayad"), []byte("1"))

			b := db.MultiBatch(b1, b2)
			Expect(db.Write(b, nil)).To(Succeed())

			Expect(sub1.Get([]byte("edmarbacha"), nil)).To(BeEquivalentTo("4"))
			Expect(sub1.Get([]byte("persioarida"), nil)).To(BeEquivalentTo("6"))
			Expect(sub1.Get([]byte("pedromalan"), nil)).To(BeEquivalentTo("5"))
			Expect(sub2.Get([]byte("persioarida"), nil)).To(BeEquivalentTo("2"))
			Expect(sub2.Get([]byte("andrelararesende"), nil)).To(BeEquivalentTo("1"))
			Expect(sub2.Get([]byte("joaosayad"), nil)).To(BeEquivalentTo("1"))
		})
	})
})
