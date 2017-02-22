package fontshot

import (
	"github.com/unixpickle/anydiff"
	"github.com/unixpickle/anynet"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/serializer"
)

func init() {
	var m Model
	serializer.RegisterTypedDeserializer(m.SerializerType(), DeserializeModel)
}

// A Model encapsulates the learner and the classifier.
type Model struct {
	// Learner takes an example as input and produces a
	// knowledge vector as output.
	// Knowledge vectors are averaged before being fed to
	// the classifier.
	Learner anynet.Net

	// Mix learned+input pairs for Classifier.
	Mixer anynet.Mixer

	Classifier anynet.Net
}

// DeserializeModel deserializes a Model.
func DeserializeModel(d []byte) (*Model, error) {
	var m Model
	err := serializer.DeserializeAny(d, &m.Learner, &m.Mixer, &m.Classifier)
	if err != nil {
		return nil, essentials.AddCtx("deserialize Model", err)
	}
	return &m, nil
}

// Apply looks at the examples and then classifies a batch
// of new images based on the examples.
// The resulting classifications are pre-sigmoid
// probabilities.
func (m *Model) Apply(examples, inputs anydiff.Res, numExample, numInputs int) anydiff.Res {
	learnedOuts := m.Learner.Apply(examples, numExample)
	avg := anydiff.SumRows(&anydiff.Matrix{
		Data: learnedOuts,
		Rows: numExample,
		Cols: learnedOuts.Output().Len() / numExample,
	})

	c := avg.Output().Creator()
	zeros := anydiff.NewConst(c.MakeVector(avg.Output().Len() * numInputs))
	repAvg := anydiff.AddRepeated(zeros, avg)

	mixed := m.Mixer.Mix(repAvg, inputs, numInputs)
	return m.Classifier.Apply(mixed, numInputs)
}

// SerializerType returns the unique ID used to serialize
// a Model with the serializer package.
func (m *Model) SerializerType() string {
	return "github.com/unixpickle/fontshot.Model"
}

// Serialize serializes the model.
func (m *Model) Serialize() (d []byte, err error) {
	defer essentials.AddCtxTo("serialize Model", &err)
	return serializer.SerializeAny(m.Learner, m.Mixer, m.Classifier)
}