//
// Copyright 2020-2022 Sean C Foley
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package ipaddr

// ExtendedIdentifierString is a common interface for strings that identify hosts, namely [IPAddressString], [MACAddressString], and [HostName].
type ExtendedIdentifierString interface {
	HostIdentifierString

	// GetAddress returns the identified address or nil if none.
	GetAddress() AddressType

	// ToAddress returns the identified address or an error.
	ToAddress() (AddressType, error)

	// Unwrap returns the wrapped [IPAddressString], [MACAddressString] or [HostName] as an interface, HostIdentifierString.
	Unwrap() HostIdentifierString
}

// WrappedIPAddressString wraps an IPAddressString to get an ExtendedIdentifierString, an extended polymorphic type.
type WrappedIPAddressString struct {
	*IPAddressString
}

// Unwrap returns the wrapped IPAddressString as an interface, HostIdentifierString.
func (w WrappedIPAddressString) Unwrap() HostIdentifierString {
	res := w.IPAddressString
	if res == nil {
		return nil
	}
	return res
}

// ToAddress returns the identified address or an error.
func (w WrappedIPAddressString) ToAddress() (AddressType, error) {
	addr, err := w.IPAddressString.ToAddress()
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// GetAddress returns the identified address or nil if none.
func (w WrappedIPAddressString) GetAddress() AddressType {
	if addr := w.IPAddressString.GetAddress(); addr != nil {
		return addr
	}
	return nil
}

// WrappedMACAddressString wraps a MACAddressString to get an ExtendedIdentifierString.
type WrappedMACAddressString struct {
	*MACAddressString
}

// Unwrap returns the wrapped MACAddressString as an interface, HostIdentifierString.
func (w WrappedMACAddressString) Unwrap() HostIdentifierString {
	res := w.MACAddressString
	if res == nil {
		return nil
	}
	return res
}

// ToAddress returns the identified address or an error.
func (w WrappedMACAddressString) ToAddress() (AddressType, error) {
	addr, err := w.MACAddressString.ToAddress()
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// GetAddress returns the identified address or nil if none.
func (w WrappedMACAddressString) GetAddress() AddressType {
	if addr := w.MACAddressString.GetAddress(); addr != nil {
		return addr
	}
	return nil
}

// WrappedHostName wraps a HostName to get an ExtendedIdentifierString.
type WrappedHostName struct {
	*HostName
}

// Unwrap returns the wrapped HostName as an interface, HostIdentifierString.
func (w WrappedHostName) Unwrap() HostIdentifierString {
	res := w.HostName
	if res == nil {
		return nil
	}
	return res
}

// ToAddress returns the identified address or an error.
func (w WrappedHostName) ToAddress() (AddressType, error) {
	addr, err := w.HostName.ToAddress()
	if err != nil {
		return nil, err
	}
	return addr, nil
}

// GetAddress returns the identified address or nil if none.
func (w WrappedHostName) GetAddress() AddressType {
	if addr := w.HostName.GetAddress(); addr != nil {
		return addr
	}
	return nil
}

var (
	_, _, _ ExtendedIdentifierString = WrappedIPAddressString{}, WrappedMACAddressString{}, WrappedHostName{}
)
