/*
Copyright (c) 2007 John Dyer (http://johndyer.name)
MIT style license
*/

if (!window.Refresh) Refresh = {};
if (!Refresh.Web) Refresh.Web = {};

Refresh.Web.SlidersList = [];

Refresh.Web.Slider = new Class({
	_bar: null,
	_arrow: null,

    Implements: [Options],

    options: {
        xMinValue: 0,
        xMaxValue: 100,
        yMinValue: 0,
        yMaxValue: 100,
		arrowCharacter: '&#8226;',
		arrowSize: '20px'
    },

	initialize: function(id, options) {
	
		this.id = id;
        this.setOptions(options);

		this.xValue = 0;
		this.yValue = 0;

		// hook up controls
		this._bar = $(this.id);

		// build controls
		this._arrow = new Element('span', {
			class: 'colorpicker-arrow',
			html: this.options.arrowCharacter
		});
		this._arrow.style.fontSize = this.options.arrowSize;
        this._arrow.inject(document.body);
        this.Arrow = this._arrow;

		// attach 'this' to html objects
		var slider = this;
		
		this.setPositioningVariables();
		
		var _this = this;
		this._event_docMouseMove = function (e) { _this._docMouseMove.call(_this, e); };
		this._event_docMouseUp   = function (e) { _this._docMouseUp.call(_this, e); };

		this._bar.addEvent('mousedown', function (e) { _this._bar_mouseDown.call(_this, e); });
		this._arrow.addEvent('mousedown', function (e) { _this._arrow_mouseDown.call(_this, e); });

		// set initial position
		this.setArrowPositionFromValues();

		// fire events
		if(this.onValuesChanged)
			this.onValuesChanged(this);

		// final setup
		Refresh.Web.SlidersList.push(this);
	},
	
	
	setPositioningVariables: function() {
		// calculate sizes and ranges
		// BAR

		this._barWidth = this._bar.getWidth();
		this._barHeight = this._bar.getHeight();
		
		var pos = this._bar.getPosition(); // FIXME: May not be the right kind of position...
		this._barTop = pos.y;
		this._barLeft = pos.x;
		
		this._barBottom = this._barTop + this._barHeight;
		this._barRight = this._barLeft + this._barWidth;

		// ARROW
		this._arrow = $(this._arrow);
		this._arrowWidth = this._arrow.getWidth();
		this._arrowHeight = this._arrow.getHeight();

		// MIN & MAX
		this.MinX = this._barLeft;
		this.MinY = this._barTop;

		this.MaxX = this._barRight;
		this.MinY = this._barBottom;
	},
	
	setArrowPositionFromValues: function(e) {
		this.setPositioningVariables();
		
		// sets the arrow position from XValue and YValue properties

		var arrowOffsetX = 0;
		var arrowOffsetY = 0;
		
		// X Value/Position
		if (this.options.xMinValue != this.options.xMaxValue) {

			if (this.xValue == this.options.xMinValue) {
				arrowOffsetX = 0;
			} else if (this.xValue == this.options.xMaxValue) {
				arrowOffsetX = this._barWidth-1;
			} else {

				var xMax = this.options.xMaxValue;
				if (this.options.xMinValue < 1)  {
					xMax = xMax + Math.abs(this.options.xMinValue) + 1;
				}
				var xValue = this.xValue;

				if (this.xValue < 1) xValue = xValue + 1;

				arrowOffsetX = xValue / xMax * this._barWidth;

				if (parseInt(arrowOffsetX) == (xMax-1))
					arrowOffsetX=xMax;
				else
					arrowOffsetX=parseInt(arrowOffsetX);

				// shift back to normal values
				if (this.options.xMinValue < 1)  {
					arrowOffsetX = arrowOffsetX - Math.abs(this.options.xMinValue) - 1;
				}
			}
		}
		
		// X Value/Position
		if (this.options.yMinValue != this.options.yMaxValue) {
			
			if (this.yValue == this.options.yMinValue) {
				arrowOffsetY = 0;
			} else if (this.yValue == this.options.yMaxValue) {
				arrowOffsetY = this._barHeight-1;
			} else {
			
				var yMax = this.options.yMaxValue;
				if (this.options.yMinValue < 1)  {
					yMax = yMax + Math.abs(this.options.yMinValue) + 1;
				}

				var yValue = this.yValue;

				if (this.yValue < 1) yValue = yValue + 1;

				var arrowOffsetY = yValue / yMax * this._barHeight;

				if (parseInt(arrowOffsetY) == (yMax-1))
					arrowOffsetY=yMax;
				else
					arrowOffsetY=parseInt(arrowOffsetY);

				if (this.options.yMinValue < 1)  {
					arrowOffsetY = arrowOffsetY - Math.abs(this.options.yMinValue) - 1;
				}
			}
		}

		this._setArrowPosition(arrowOffsetX, arrowOffsetY);

	},
	_setArrowPosition: function(offsetX, offsetY) {
		
		
		// validate
		if (offsetX < 0) offsetX = 0
		if (offsetX > this._barWidth) offsetX = this._barWidth;
		if (offsetY < 0) offsetY = 0
		if (offsetY > this._barHeight) offsetY = this._barHeight;

		var posX = this._barLeft + offsetX;
		var posY = this._barTop + offsetY;

		// check if the arrow is bigger than the bar area
		if (this._arrowWidth > this._barWidth) {
			posX = posX - (this._arrowWidth/2 - this._barWidth/2);
		} else {
			posX = posX - parseInt(this._arrowWidth/2);
		}
		if (this._arrowHeight > this._barHeight) {
			posY = posY - (this._arrowHeight/2 - this._barHeight/2);
		} else {
			posY = posY - parseInt(this._arrowHeight/2);
		}
		this._arrow.style.left = posX + 'px';
		this._arrow.style.top = posY + 'px';
	},
	_bar_mouseDown: function(e) {
		this._mouseDown(e);
	},
	
	_arrow_mouseDown: function(e) {
		this._mouseDown(e);
	},
	
	_mouseDown: function(e) {
		Refresh.Web.ActiveSlider = this;
		
		this.setValuesFromMousePosition(e);
		
		document.addEvent('mousemove', this._event_docMouseMove);
		document.addEvent('mouseup', this._event_docMouseUp);

		e.stop();
	},
	
	_docMouseMove: function(e) {
		this.setValuesFromMousePosition(e);
		
		e.stop();
	},
	
	_docMouseUp: function(e) {
		document.removeEvent('mouseup', this._event_docMouseUp);
		document.removeEvent('mousemove', this._event_docMouseMove);
		e.stop();
	},
	
	setValuesFromMousePosition: function(e) {
		//this.setPositioningVariables();
		
		var mouse = e.page;
		
		var relativeX = 0;
		var relativeY = 0;

		// mouse relative to object's top left
		if (mouse.x < this._barLeft)
			relativeX = 0;
		else if (mouse.x > this._barRight)
			relativeX = this._barWidth;
		else
			relativeX = mouse.x - this._barLeft + 1;

		if (mouse.y < this._barTop)
			relativeY = 0;
		else if (mouse.y > this._barBottom)
			relativeY = this._barHeight;
		else
			relativeY = mouse.y - this._barTop + 1;
			

		var newXValue = parseInt(relativeX / this._barWidth * this.options.xMaxValue);
		var newYValue = parseInt(relativeY / this._barHeight * this.options.yMaxValue);
		
		// set values
		this.xValue = newXValue;
		this.yValue = newYValue;

		// position arrow
		if (this.options.xMaxValue == this.options.xMinValue)
			relativeX = 0;
		if (this.options.yMaxValue == this.options.yMinValue)
			relativeY = 0;
		this._setArrowPosition(relativeX, relativeY);

		// fire events
		if(this.onValuesChanged)
			this.onValuesChanged(this);
	}

});
