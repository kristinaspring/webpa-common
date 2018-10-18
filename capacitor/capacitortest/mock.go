package capacitortest

import (
	"github.com/Comcast/webpa-common/capacitor"
	"github.com/stretchr/testify/mock"
)

type Mock struct {
	mock.Mock
}

var _ capacitor.Interface = (*Mock)(nil)

func (m *Mock) Submit(f func()) {
	m.Called(f)
}

func (m *Mock) OnSubmit(f func()) *mock.Call {
	return m.On("Submit", f)
}

func (m *Mock) Discharge() {
	m.Called()
}

func (m *Mock) OnDischarge() *mock.Call {
	return m.On("Discharge")
}

func (m *Mock) Cancel() {
	m.Called()
}

func (m *Mock) OnCancel() *mock.Call {
	return m.On("Cancel")
}
