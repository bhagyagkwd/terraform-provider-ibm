// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// VolumeAttachmentReference VolumeAttachmentVolumeReference
// swagger:model volume_attachment_reference
type VolumeAttachmentReference struct {
	ResourceReference

	// volume
	Volume *ResourceReference `json:"volume,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *VolumeAttachmentReference) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 ResourceReference
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.ResourceReference = aO0

	// AO1
	var dataAO1 struct {
		Volume *ResourceReference `json:"volume,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.Volume = dataAO1.Volume

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m VolumeAttachmentReference) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.ResourceReference)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)

	var dataAO1 struct {
		Volume *ResourceReference `json:"volume,omitempty"`
	}

	dataAO1.Volume = m.Volume

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)

	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this volume attachment reference
func (m *VolumeAttachmentReference) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with ResourceReference
	if err := m.ResourceReference.Validate(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVolume(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *VolumeAttachmentReference) validateVolume(formats strfmt.Registry) error {

	if swag.IsZero(m.Volume) { // not required
		return nil
	}

	if m.Volume != nil {
		if err := m.Volume.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("volume")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *VolumeAttachmentReference) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VolumeAttachmentReference) UnmarshalBinary(b []byte) error {
	var res VolumeAttachmentReference
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}