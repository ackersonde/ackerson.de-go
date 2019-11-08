jQuery( function ( $ ) {
	
	'use strict';
	
	if ( $('div.overlay-container').length ) {
		
		// Find all iframes inside the overlays
		var $doviframes = $( "#sidebar-overlay .overlay iframe" );
		
		setTimeout( function() { 
		
			$doviframes.each( function() {
				
				var iframeHeight = this.height;
				
				if ( iframeHeight == '' ) {
					
					iframeHeight = $( this ).height();
				}
				
				var iframeWidth = this.width;
				
				if ( iframeWidth == '' ) {
					
					iframeWidth = $( this ).width();
				}
				
				iframeHeight = parseInt( iframeHeight );
				iframeWidth = parseInt( iframeWidth );
				
				var ratio = iframeHeight / iframeWidth;
				
				$( this ).attr( "data-ratio", ratio )
				// Remove the hardcoded width & height attributes
				.removeAttr( "width" )
				.removeAttr( "height" );
				
				// Get the parent container's width
				var width = $( this ).parent().width();
				
				$( this ).width( width ).height( width * ratio );
			});
			
		}, 200);
		
		// Resize the iframes when the window is resized
		$( window ).resize( function () {
		  $doviframes.each( function() {
			// Get the parent container's width
			var width = $( this ).parent().width();
			$( this ).width( width )
			  .height( width * $( this ).data( "ratio" ) );
		  });
		});
		
		// Divi Cascade Fix
		if ( $('#et-builder-module-design-cached-inline-styles').length && !$('#cloned-et-builder-module-design-cached-inline-styles').length ) {
			
			// Divi Cached Inline Styles
			var divicachedcss = $( '#et-builder-module-design-cached-inline-styles' );
			var divicachedcsscontent = divicachedcss.html();
			
			// Remove #page-container from Divi Cached Inline Styles tag and cloning it to prevent issues
			divicachedcsscontent = divicachedcsscontent.replace(/\#page-container/g, ' ');
			
			$( divicachedcss ).after( '<style id="cloned-et-builder-module-design-cached-inline-styles">' + divicachedcsscontent + '</style>' );
		}
		
		$(document).keyup(function(e) {
			
			if (e.keyCode == 27) {
				
				closeActiveOverlay();
			}
		});
		
		$(window).load(function() {
			
			if ( window.location.hash ) {
				
				var hash = window.location.hash.substring( 1 );
				var idx_overlay = hash.indexOf('overlay'); 
				
				if ( idx_overlay !== -1 ) {
					
					var idx_overlayArr = hash.split('-');
					
					if ( idx_overlayArr.length > 1 ) {
						
						var overlay_id = idx_overlayArr[1];
						
						showOverlay( overlay_id );
					}
				}
			}
		});
		
		var overlay_container = $( 'div.overlay-container' );
		var container = $( 'div#page-container' );
		
		// Remove any duplicated overlay
		$( overlay_container ).each(function () {
			$('[id="' + this.id + '"]:gt(0)').remove();
		});
		
		$('body').on( 'click touch tap', '[id^="overlay_"]', function(e) {
			
			var overlayArr = $(this).attr('id').split('_')
			, overlay_id = overlayArr[3]
			, cookieName = 'divioverlay' + overlay_id;
			
			doEraseCookie( cookieName );
			
			showOverlay( overlay_id );
		});
		
		$('body').on( 'click touch tap', '[rel^="unique_overlay_"]', function(e) {
			
			var overlayArr = $(this).attr('rel').split('_')
			, overlay_id = overlayArr[4]
			, cookieName = 'divioverlay' + overlay_id;
			
			doEraseCookie( cookieName );
			
			showOverlay( overlay_id );
			
			e.preventDefault();
		});
		
		$('body').on( 'click touch tap', '[class*="divioverlay-"], [class*="overlay-"]', function(e) {
			
			var overlayArr = $(this).attr('class').split(' ');
			
			$( overlayArr ).each(function( index,value ) {
				
				var idx_overlay = value.indexOf('overlay');
				
				if ( idx_overlay !== -1 ) {
					
					var idx_overlayArr = value.split('-');
					
					if ( idx_overlayArr.length > 1 ) {
						
						var overlay_id = idx_overlayArr[1]
						, cookieName = 'divioverlay' + overlay_id;
						
						doEraseCookie( cookieName );
						
						showOverlay( overlay_id );
					}
				}
			});
		});
		
		if (typeof overlays_with_css_trigger !== 'undefined') {
			
			if ( $( overlays_with_css_trigger ).length > 0 ) {
				
				$.each( overlays_with_css_trigger, function( overlay_id, selector ) {
					
					$( selector ).on('click touch tap', function () {
						
						var cookieName = 'divioverlay' + overlay_id;
						
						doEraseCookie( cookieName );
						
						showOverlay( overlay_id );
					});
				});
			}
		}
		
		if (typeof overlays_with_automatic_trigger !== 'undefined') {
			
			if ( $( overlays_with_automatic_trigger ).length > 0 ) {
				
				$.each( overlays_with_automatic_trigger, function( overlay_id, at_settings ) {
					
					var at_settings_parsed = jQuery.parseJSON( at_settings );
					var at_type_value = at_settings_parsed.at_type;
					var at_onceperload = at_settings_parsed.at_onceperload;
					
					if ( at_onceperload == 1 ) {
						
						showOverlayOnce( overlay_id );
					}
					
					if ( at_type_value == 'overlay-timed' ) {
						
						var time_delayed = at_settings_parsed.at_value * 1000;
                        
                        if ( time_delayed == 0 ) {
                         
                            time_delayed = 1000;
                        }
						
						setTimeout( function() {
							
							showOverlay( overlay_id );
							
						}, time_delayed);
					}
					
					
					if ( at_type_value == 'overlay-scroll' ) {
						
						var overlayScroll = at_settings_parsed.at_value, refScroll;
						
						if ( overlayScroll.indexOf('%') || overlayScroll.indexOf('px') ) {
							
							if ( overlayScroll.indexOf('%') ) {
								
								overlayScroll = overlayScroll.replace(/%/g, '');
								refScroll = '%';
							}
							
							if ( overlayScroll.indexOf('px') ) {
								
								overlayScroll = overlayScroll.replace(/px/g, '');
								refScroll = 'px';
							}
							
							overlayScroll = overlayScroll.split(':');
							var overlayScrollFrom = overlayScroll[0];
							var overlayScrollTo = overlayScroll[1];
							
							$(window).scroll(function(e) {
							
								var s = getScrollTop(),
									d = $(document).height(),
									c = $(window).height(),
									wScroll;
								
								if ( refScroll == '%' ) {
									
									wScroll = (s / (d-c)) * 100;
								
								} else if ( refScroll == 'px' ) {
									
									wScroll = s;
									
								} else {
									
									return;
								}
								
								if ( overlayScrollFrom > 0 && overlayScrollTo > 0 ) {
									
									if ( overlayScrollFrom <= wScroll && overlayScrollTo >= wScroll ) {
										
										if ( !isActiveOverlay( overlay_id ) ) {
											
											showOverlay( overlay_id );
										}
									}
									else if ( isActiveOverlay( overlay_id ) ) {
										
										closeActiveOverlay( overlay_id );
									}
								}
								
								if ( overlayScrollFrom > 0 && overlayScrollTo == '' ) {
									
									if ( overlayScrollFrom <= wScroll ) {
										
										if ( !isActiveOverlay( overlay_id ) ) {
											
											showOverlay( overlay_id );
										}
									}
									else if ( isActiveOverlay( overlay_id ) ) {
										
										closeActiveOverlay( overlay_id );
									}
								}
								
								if ( overlayScrollFrom == '' && overlayScrollTo > 0 ) {
									
									if ( overlayScrollTo >= wScroll ) {
										
										if ( !isActiveOverlay( overlay_id ) ) {
											
											showOverlay( overlay_id );
										}
									}
									else if ( isActiveOverlay( overlay_id ) ) {
									
										closeActiveOverlay( overlay_id );
									}
								}
							});
						}
					}
					
					
					if ( at_type_value == 'overlay-exit' ) {
						
						$.exitIntent('enable', { 'sensitivity': 100 });
						
						$(document).bind('exitintent',
							function() {
								
								var overlay = '#overlay-' + overlay_id
								, at_onceperload = $( overlay ).attr( 'data-displayonceperload' )
								, displayonceperloadpassed = $( overlay ).attr( 'data-displayonceperloadpassed' );
								
								if ( !isActiveOverlay( overlay_id ) ) {
									
									if ( at_onceperload == 1 && displayonceperloadpassed != 1 ) {
										
										showOverlay( overlay_id );
										
										$( overlay ).attr( 'data-displayonceperloadpassed', 1 );
									}
									
									if ( undefined === at_onceperload ) {
										
										showOverlay( overlay_id );
									}
								}
							});
					}
				});
			}
		}
		
		
		$('.nav a, .mobile_nav a').each(function( index,value ) {
			
			var href = $( value ).attr('href');
			
			if ( href !== undefined ) {
			
				var idx_overlay = href.indexOf('overlay');
				
				if ( idx_overlay !== -1 ) {
					
					var idx_overlayArr = href.split('-');
					
					if ( idx_overlayArr.length > 1 ) {
						
						var overlay_id = idx_overlayArr[1];
						
						$(this).attr('data-overlayid', overlay_id);
						
						$(this).on('click touch tap', function () {
							
							overlay_id = $(this).data('overlayid');
							
							showOverlay( overlay_id );
						});
					}
				}
			}
		});
		
		var transEndEventNames = {
			'WebkitTransition': 'webkitTransitionEnd',
			'MozTransition': 'transitionend',
			'OTransition': 'oTransitionEnd',
			'msTransition': 'MSTransitionEnd',
			'transition': 'transitionend'
		},
		transEndEventName = transEndEventNames[ Modernizr.prefixed( 'transition' ) ],
		support = { transitions : Modernizr.csstransitions };
		
		function shuffle(array) {
			var currentIndex = array.length
			, temporaryValue
			, randomIndex
			;

			// While there remain elements to shuffle...
			while (0 !== currentIndex) {
				// Pick a remaining element...
				randomIndex = Math.floor(Math.random() * currentIndex);
				currentIndex -= 1;
				// And swap it with the current element.
				temporaryValue = array[currentIndex];
				array[currentIndex] = array[randomIndex];
				array[randomIndex] = temporaryValue;
			}
			return array;
		}
		
		function toggleOverlay( overlay_id ) {
			
			var overlay_selector = '#overlay-' + overlay_id;
			var overlay_cache_selector = '#overlay-' + overlay_id;
			var overlay_container = '#divi-overlay-container-' + overlay_id;
			var overlay = $('body').find( overlay_cache_selector );
			var oid = overlay.attr('id');
			var prevent_mainpage_scroll = overlay.data('preventscroll');
			var displayonceperload = overlay.data('displayonceperload');
			var overlay_active_selector = 'div.overlay-container div.overlay-body';
			var preventOpen = overlay.attr('data-preventopen');
			
			var cookieName = 'divioverlay' + overlay_id
			, cookieDays = overlay.data('cookie');
			
			if ( doReadCookie( cookieName ) && cookieDays != 0 ) {
				
				return;
			}
			
			if ( $( overlay_cache_selector ).hasClass("overlay-cornershape") ) {
				
				var s = Snap( document.querySelector( overlay_cache_selector ).querySelector( 'svg' ) );
				var original_s = Snap( document.querySelector( overlay_selector ).querySelector( 'svg' ) );
				
				var path = s.select( 'path' );
				var original_path = original_s.select( 'path' );
				
				var pathConfig = {
					from : original_path.attr( 'd' ),
					to : document.querySelector( overlay_cache_selector ).getAttribute( 'data-path-to' )
				};
			}
			
			if ( $( overlay_cache_selector ).hasClass("overlay-boxes") ) {
				
				paths = [].slice.call( document.querySelector( overlay_cache_selector ).querySelectorAll( 'svg > path' ) ),
				pathsTotal = paths.length;
			}
			
			if ( $( overlay_cache_selector ).hasClass("overlay-genie") ) {
				
				var s1 = Snap( document.querySelector( overlay_cache_selector ).querySelector( 'svg' ) ), 
				path1 = s1.select( 'path' ),
				steps = document.querySelector( overlay_cache_selector ).getAttribute( 'data-steps' ).split(';'),
				stepsTotal = steps.length;
			}
			
			if ( $( overlay_cache_selector ).hasClass('overlay-boxes') ) {
				
				var cnt = 0;
				
				shuffle( paths );
			}
			
			
			// Add Divi Builder main section class after others plugins are ready
			if ( ! $( overlay ).find('.entry-content > div').hasClass('et_pb_section') ) {
				
				$( overlay ).find('.entry-content > div').addClass('et_pb_section');
			}
			
			
			if ( $( overlay ).hasClass('open') ) {
				
				if ( cookieDays > 0 ) {
					
					doCreateCookie( cookieName, 'true', cookieDays );
				}
				
				$( overlay ).removeClass('open');
				$( overlay ).addClass('close');
				
				if ( $( overlay_cache_selector ).hasClass('overlay-hugeinc') 
					|| $( overlay_cache_selector ).hasClass('overlay-corner') 
					|| $( overlay_cache_selector ).hasClass('overlay-scale') ) {
						
					$( overlay_cache_selector ).css('opacity',0);
				}
				
				if ( $( overlay_cache_selector ).hasClass('overlay-contentpush') ) {
					
					$( container ).removeClass('overlay-contentpush-open');
					
					setTimeout( function() { 
					
						$( container ).removeClass( 'container2' );
						$("html,body").removeAttr('style'); 
						
					}, 1000);
				}
				
				if ( $( overlay_cache_selector).hasClass('overlay-contentscale') ) {
					
					$( container ).removeClass('overlay-contentscale-open');
					
					setTimeout( function() {
						
						$( container ).removeClass( 'container3' );
						
					}, 1000);
				}
				
				if ( $( overlay_cache_selector ).hasClass('overlay-cornershape') ) {
					
					var onEndTransitionFn = function( ev ) {
						
						$( overlay ).removeClass( 'close' );
					};
					path.animate( { 'path' : pathConfig.from }, 400, mina.linear, onEndTransitionFn );
				}
				else if ( $( overlay_cache_selector ).hasClass('overlay-boxes') ) {
					
					paths.forEach( function( p, i ) {
						setTimeout( function() {
							++cnt;
							p.style.display = 'none';
							if( cnt === pathsTotal ) {
								
								$( overlay ).removeClass( 'close' );
							}
						}, i * 30 );
					});
				}
				else if ( $(  overlay_cache_selector ).hasClass('overlay-genie') ) {
					
					var pos = stepsTotal-1;
					var onEndTransitionFn = function( ev ) {
						
						$( overlay ).removeClass( 'close' );
					},
					nextStep = function( pos ) {
						pos--;
						if( pos < 0 ) return;
						path1.animate( { 'path' : steps[pos] }, 60, mina.linear, function() { 
							if( pos === 0 ) {
								onEndTransitionFn();
							}
							nextStep( pos );
						} );
					};

					nextStep( pos );
				}
				else {
					
					overlay = document.querySelector( overlay_cache_selector );
					
					var onEndTransitionFn = function( ev ) {
						if( support.transitions ) {
							if( ev.propertyName !== 'visibility' ) return;
							this.removeEventListener( transEndEventName, onEndTransitionFn );
						}
						
						$( overlay ).removeClass( 'close' );
					};
					
					if ( support.transitions ) {
						
						overlay.addEventListener( transEndEventName, onEndTransitionFn );
					}
					else {
						
						onEndTransitionFn();
					}
				}
				
				if ( prevent_mainpage_scroll ) {
					
					$( 'html,body' ).removeClass('prevent_mainpage_scroll');
					$( 'body' ).removeClass('prevent_mainpage_scroll_mobile');
					$( '#page-container' ).removeClass('prevent_content_scroll');
					$('html, body').scrollTop( $( overlay ).attr('data-scrolltop') );
					$( overlay ).attr('data-scrolltop', '' );
				}
				
				setTimeout( function() {
					
					$( overlay_cache_selector ).removeAttr('style');
					$( overlay_cache_selector + ' path' ).removeAttr('style');
					
					if ( !isActiveOverlay() ) {
						
						$( "#page-container" ).css('z-index','1');
						$( "#page-container .container" ).css('z-index','1');
						$( "#page-container #main-header" ).css('z-index','99989');
						$( "#wpadminbar" ).css('z-index','100001');
						$( "#sidebar-overlay" ).css('z-index','-15');
					}
					
					togglePlayableTags( '#overlay-' + overlay_id );
					
				}, 500);
			}
			else if( !$( overlay ).hasClass('close') ) {
				
				overlay.attr('data-scrolltop', getScrollTop() );
				
				$( "#page-container" ).css('z-index','0');
				$( "#page-container .container" ).css('z-index','0');
				$( "#page-container #main-header" ).css('z-index','-1');
				$( "#wpadminbar" ).css('z-index','0');
				$( "#sidebar-overlay" ).css('z-index','16777210');
				
				setTimeout( function() {
					
					$( overlay ).addClass('open');
					
					if ( overlay.attr('data-bgcolor') != "") {
						$( overlay_cache_selector ).css( { 'background-color': overlay.attr('data-bgcolor') } );
					}
					
					if ( overlay.attr('data-fontcolor') != "") {
						$( overlay_cache_selector ).css( 'color', overlay.attr('data-fontcolor') );
					}
					
					if ( $( overlay_cache_selector ).hasClass('overlay-contentpush') ) {
						
						$( "html,body" ).css('overflow-x','hidden');
						
						$( overlay_cache_selector ).css('opacity',1);
						
						container.attr('class', 'container2');
						
						$( container ).addClass( 'overlay-contentpush-open' );
					}
					
					if ( $( overlay_cache_selector ).hasClass('overlay-contentscale')) {
						
						container.attr('class', 'container3');
						
						$( container ).addClass('overlay-contentscale-open');
					}
					
					if ( $( overlay_cache_selector ).hasClass('overlay-cornershape')) {
						
						$( overlay_cache_selector ).css({"background":"transparent none repeat scroll 0 0"});
						
						path.animate( { 'path' : pathConfig.to }, 400, mina.linear );
						$( overlay_cache_selector + ' .overlay-path' ).css({"fill": overlay.attr('data-bgcolor')});
					}
					
					if ( $(  overlay_cache_selector ).hasClass('overlay-boxes') ) {
						
						$( overlay_cache_selector ).css({"background":"transparent none repeat scroll 0 0"});
						paths.forEach( function( p, i ) {
							setTimeout( function() {
								p.style.display = 'block';
								p.style.fill = overlay.attr('data-bgcolor');
							}, i * 30 );
						});
					}
					
					if ( $( overlay_cache_selector ).hasClass('overlay-genie') ) {
						
						$( overlay_cache_selector ).css({"background":"transparent none repeat scroll 0 0"});
						
						var pos = 0;
						
						$( overlay ).addClass( 'open' );
						
						var nextStep = function( pos ) {
							pos++;
							if( pos > stepsTotal - 1 ) return;
							path1.animate( { 'path' : steps[pos] }, 60, mina.linear, function() { nextStep(pos); } );
							
							$( overlay_cache_selector + ' .overlay-path' ).css({"fill": overlay.attr('data-bgcolor')});
						};
						
						nextStep(pos);
					}
					
					if ( prevent_mainpage_scroll ) {
						
						$( 'html,body' ).addClass('prevent_mainpage_scroll');
						$( 'body' ).addClass('prevent_mainpage_scroll_mobile');
						$( '#page-container' ).addClass('prevent_content_scroll');
					}
					
					setTimeout( function() {
						
						dov_initDiviElements
						
					}, 200);
					
				}, 200);
			}
		}
		
		
		function dov_initDiviElements() {
			
			// Set Divi Elements
			var $et_pb_circle_counter = $(".divibars-body .et_pb_circle_counter"),
				$et_pb_number_counter = $(".divibars-body .et_pb_number_counter"),
				$et_pb_countdown_timer = $(".divibars-body .et_pb_countdown_timer");
				
			$('.divibars-body .et_animated').each(function() {
				dov_et_remove_animation( $( this ) );
			});
			
			// Init Divi Elements
			setTimeout( function() {
				
				window.et_fix_testimonial_inner_width(), 
				$et_pb_circle_counter.length && window.et_pb_reinit_circle_counters($et_pb_circle_counter), 
				$et_pb_number_counter.length && window.et_pb_reinit_number_counters($et_pb_number_counter), 
				$et_pb_countdown_timer.length && window.et_pb_countdown_timer_init($et_pb_countdown_timer),
				window.et_reinit_waypoint_modules();
				
			}, 1);
		}
		
		
		function dov_et_get_animation_classes() {
			return ["et_animated", "infinite", "fade", "fadeTop", "fadeRight", "fadeBottom", "fadeLeft", "slide", "slideTop", "slideRight", "slideBottom", "slideLeft", "bounce", "bounceTop", "bounceRight", "bounceBottom", "bounceLeft", "zoom", "zoomTop", "zoomRight", "zoomBottom", "zoomLeft", "flip", "flipTop", "flipRight", "flipBottom", "flipLeft", "fold", "foldTop", "foldRight", "foldBottom", "foldLeft", "roll", "rollTop", "rollRight", "rollBottom", "rollLeft"]
		}
		
		function dov_et_remove_animation($element) {
			var animation_classes = dov_et_get_animation_classes();
			$element.removeClass(animation_classes.join(" ")), $element.removeAttr("style")
		}
		
		
		function getScrollTop() {
			
			if ( typeof pageYOffset!= 'undefined' ) {
				
				// most browsers except IE before #9
				return pageYOffset;
			}
			else {
				
				var B = document.body; // IE 'quirks'
				var D = document.documentElement; // IE with doctype
				D = ( D.clientHeight ) ? D: B;
				
				return D.scrollTop;
			}
		}
		
		
		function showOverlay( overlay_id ) {
			
			if ( !DovisInt( overlay_id ) )
				return;
			
			var data = {
				action: 'divioverlay_content',
				overlay_id: overlay_id
			};
			
			var divi_overlay_container_selector = '#divi-overlay-container-' + overlay_id;
			
			if ( $( divi_overlay_container_selector ).length ) {
			
				var overlay_body = $( divi_overlay_container_selector ).find( '.overlay' );
				
				toggleSrcInPlayableTags( overlay_body );
				
				toggleOverlay( overlay_id );
			}
		}
		
		function showOverlayOnce( overlay_id ) {
			
			if ( !DovisInt( overlay_id ) )
				return;
			
			var overlay = '#overlay-' + overlay_id;
			
			$( overlay ).attr( 'data-displayonceperload', 1 );
		}
		
		function toggleSrcInPlayableTags( str ) {
			
			str.find("iframe").each(function() { 
				var src = $(this).data('src');
				$(this).attr('src', src);  
			});
			
			return str;
		}
		
		$('body').on('click', '.overlay.open, .overlay-close, .overlay-close span, .close-divi-overlay', function(e) {
			
			if ( e.target !== e.currentTarget ) return;
			
			closeActiveOverlay();
		});
		
		function closeActiveOverlay( overlay_id ) {
			
			// find active overlay
			var overlay = $( '#sidebar-overlay' ).find('.overlay-container .overlay.open:last');
			
			var displayonceperload = overlay.data('displayonceperload');
			
			if ( overlay.length ) {
				
				if ( overlay_id == null ) {
					
					var overlayArr = overlay.attr('id').split('-');
					overlay_id = overlayArr[ overlayArr.length - 1 ];
				}
				
				showOverlay( overlay_id );
			}
		}
		
		function getActiveOverlay( onlyNumber ) {
			
			// find active overlay
			var overlay = $( 'body' ).find( '.overlay.active' );
			var overlay_id = null;
			
			if ( overlay.length ) {
				
				var overlayArr = overlay.attr('id').split('-');
				overlay_id = overlayArr[1];
			}
			
			return overlay_id;
		}
		
		function isOpeningOverlay( overlay_id ) {
			
			if ( !overlay_id ) {
				
				return null;
			}
			
			var overlay = $( '#overlay-' + overlay_id );
			
			if ( $( overlay ).css('opacity') < 1 ) {
				
				return true;
			}
			
			return false;
		}
		
		function isClosingOverlay( overlay_id ) {
			
			if ( !overlay_id ) {
				
				return null;
			}
			
			var overlay = $( '#overlay-' + overlay_id );
			
			if ( $( overlay ).hasClass('close') ) {
				
				return false;
			}
			
			return true;
		}
		
		function isActiveOverlay( overlay_id ) {
			
			if ( !overlay_id ) {
				
				var overlay = $( '.overlay.open' );
			}
			else {
				
				var overlay = $( '#overlay-' + overlay_id );
			}
			
			if ( $( overlay ).hasClass('open') ) {
				
				return true;
			}
			
			return false;
		}
		
		function doCreateCookie( name,value,days ) {
			
			var expires = "";
			
			if ( days ) {
				
				var date = new Date();
				
				date.setTime(date.getTime() + ( days * 24 * 60 * 60 * 1000));
				
				expires = "; expires=" + date.toUTCString();
			}
			
			document.cookie = name + "=" + value + expires + "; path=/";
		}
		
		function doReadCookie( name ) {
			
			var nameEQ = name + "=";
			var ca = document.cookie.split(';');
			
			for(var i=0;i < ca.length;i++) {
				
				var c = ca[i];
				
				while (c.charAt(0)==' ') c = c.substring(1,c.length);
				
				if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
			}
			
			return null;
		}
		
		function doEraseCookie( name ) {
			doCreateCookie( name, '', -1 );
		}
	}
	
});


function DovisInt(value) {
    var x;
    return isNaN(value) ? !1 : (x = parseFloat(value), (0 | x) === x);
}