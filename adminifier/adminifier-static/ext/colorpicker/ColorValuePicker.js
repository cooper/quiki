/*
Copyright (c) 2007 John Dyer (http://johndyer.name)
MIT style license
*/

if (!window.Refresh) Refresh = {};
if (!Refresh.Web) Refresh.Web = {};

Refresh.Web.ColorValuePicker = new Class({
	initialize: function(id) {
		this.id = id;

		this.onValuesChanged = null;

		this._hueInput = $(this.id + '-hue');
		this._valueInput = $(this.id + '-brightness');
		this._saturationInput = $(this.id + '-saturation');

		this._redInput = $(this.id + '-red');
		this._greenInput = $(this.id + '-green');
		this._blueInput = $(this.id + '-blue');

		this._hexInput = $(this.id + '-hex');

		// assign events

		// events
		var _this = this;
		this._event_onHsvKeyUp 	= function (e) { _this._onHsvKeyUp.call(_this, e); };
		this._event_onHsvBlur 	= function (e) { _this._onHsvBlur.call(_this, e); };
		this._event_onRgbKeyUp 	= function (e) { _this._onRgbKeyUp.call(_this, e); };
		this._event_onRgbBlur 	= function (e) { _this._onRgbBlur.call(_this, e); };
		this._event_onHexKeyUp	= function (e) { _this._onHexKeyUp.call(_this, e); };
		
		// HSB
		this._hueInput.addEvent('keyup', this._event_onHsvKeyUp);
		this._valueInput.addEvent('keyup',this._event_onHsvKeyUp);
		this._saturationInput.addEvent('keyup',this._event_onHsvKeyUp);
		this._hueInput.addEvent('blur', this._event_onHsvBlur);
		this._valueInput.addEvent('blur',this._event_onHsvBlur);
		this._saturationInput.addEvent('blur',this._event_onHsvBlur);

		// RGB
		this._redInput.addEvent('keyup', this._event_onRgbKeyUp);
		this._greenInput.addEvent('keyup', this._event_onRgbKeyUp);
		this._blueInput.addEvent('keyup', this._event_onRgbKeyUp);
		this._redInput.addEvent('blur', this._event_onRgbBlur);
		this._greenInput.addEvent('blur', this._event_onRgbBlur);
		this._blueInput.addEvent('blur', this._event_onRgbBlur);

		// HEX
		this._hexInput.addEvent('keyup', this._event_onHexKeyUp);
		
		this.color = new Refresh.Web.Color();
		
		// get an initial value
		if (this._hexInput.value != '')
			this.color.setHex(this._hexInput.value);
			
		// set the others based on initial value
		this._hexInput.value = this.color.hex;
		
		this._redInput.value = this.color.r;
		this._greenInput.value = this.color.g;
		this._blueInput.value = this.color.b;
		
		this._hueInput.value = this.color.h;
		this._saturationInput.value = this.color.s;
		this._valueInput.value = this.color.v;

	},
	_onHsvKeyUp: function(e) {
		if (e.target.value == '') return;
		this.validateHsv(e);
		this.setValuesFromHsv();
		if (this.onValuesChanged) this.onValuesChanged(this);
	},
	_onRgbKeyUp: function(e) {
		if (e.target.value == '') return;
		this.validateRgb(e);
		this.setValuesFromRgb();
		if (this.onValuesChanged) this.onValuesChanged(this);
	},
	_onHexKeyUp: function(e) {
		if (e.target.value == '') return;
		this.validateHex(e);
		this.setValuesFromHex();
		if (this.onValuesChanged) this.onValuesChanged(this);
	},
	_onHsvBlur: function(e) {
		if (e.target.value == '')
			this.setValuesFromRgb();
	},
	_onRgbBlur: function(e) {
		if (e.target.value == '')
			this.setValuesFromHsv();
	},
	HexBlur: function(e) {
		if (e.target.value == '')
			this.setValuesFromHsv();
	},
	validateRgb: function(e) {
		if (!this._keyNeedsValidation(e)) return e;
		this._redInput.value = this._setValueInRange(this._redInput.value,0,255);
		this._greenInput.value = this._setValueInRange(this._greenInput.value,0,255);
		this._blueInput.value = this._setValueInRange(this._blueInput.value,0,255);
	},
	validateHsv: function(e) {
		if (!this._keyNeedsValidation(e)) return e;
		this._hueInput.value = this._setValueInRange(this._hueInput.value,0,359);
		this._saturationInput.value = this._setValueInRange(this._saturationInput.value,0,100);
		this._valueInput.value = this._setValueInRange(this._valueInput.value,0,100);
	},
	validateHex: function(e) {
		if (!this._keyNeedsValidation(e)) return e;
		var hex = new String(this._hexInput.value).toUpperCase();
		hex = hex.replace(/[^A-F0-9]/g, '0');
		if (hex.length > 6) hex = hex.substring(0, 6);
		this._hexInput.value = hex;
	},
	_keyNeedsValidation: function(e) {

		if (e.keyCode == 9  || // TAB
			e.keyCode == 16  || // Shift
			e.keyCode == 38 || // Up arrow
			e.keyCode == 29 || // Right arrow
			e.keyCode == 40 || // Down arrow
			e.keyCode == 37    // Left arrow
			||
			(e.ctrlKey && (e.keyCode == 'c'.charCodeAt() || e.keyCode == 'v'.charCodeAt()) )
		) return false;

		return true;
	},
	_setValueInRange: function(value,min,max) {
		if (value == '' || isNaN(value))
			return min;
		
		value = parseInt(value);
		if (value > max)
			return max;
		if (value < min)
			return min;
		
		return value;
	},
	setValuesFromRgb: function() {
		this.color.setRgb(this._redInput.value, this._greenInput.value, this._blueInput.value);
		this._hexInput.value = this.color.hex;
		this._hueInput.value = this.color.h;
		this._saturationInput.value = this.color.s;
		this._valueInput.value = this.color.v;
	},
	setValuesFromHsv: function() {
		this.color.setHsv(this._hueInput.value, this._saturationInput.value, this._valueInput.value);
		
		this._hexInput.value = this.color.hex;
		this._redInput.value = this.color.r;
		this._greenInput.value = this.color.g;
		this._blueInput.value = this.color.b;
	},
	setValuesFromHex: function() {
		this.color.setHex(this._hexInput.value);

		this._redInput.value = this.color.r;
		this._greenInput.value = this.color.g;
		this._blueInput.value = this.color.b;
		
		this._hueInput.value = this.color.h;
		this._saturationInput.value = this.color.s;
		this._valueInput.value = this.color.v;
	}
});
