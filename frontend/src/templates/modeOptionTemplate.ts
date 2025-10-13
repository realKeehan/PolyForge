import runtime from 'pug-runtime';
function pug_attr(t,e,n,r){if(!1===e||null==e||!e&&("class"===t||"style"===t))return"";if(!0===e)return" "+(r?t:t+'="'+t+'"');var f=typeof e;return"object"!==f&&"function"!==f||"function"!=typeof e.toJSON||(e=e.toJSON()),"string"==typeof e||(e=JSON.stringify(e),n||-1===e.indexOf('"'))?(n&&(e=pug_escape(e))," "+t+'="'+e+'"'):" "+t+"='"+e.replace(/'/g,"&#39;")+"'"}
function pug_classes(s,r){return Array.isArray(s)?pug_classes_array(s,r):s&&"object"==typeof s?pug_classes_object(s):s||""}
function pug_classes_array(r,a){for(var s,e="",u="",c=Array.isArray(a),g=0;g<r.length;g++)(s=pug_classes(r[g]))&&(c&&a[g]&&(s=pug_escape(s)),e=e+u+s,u=" ");return e}
function pug_classes_object(r){var a="",n="";for(var o in r)o&&r[o]&&pug_has_own_property.call(r,o)&&(a=a+n+o,n=" ");return a}
function pug_escape(e){var a=""+e,t=pug_match_html.exec(a);if(!t)return e;var r,c,n,s="";for(r=t.index,c=0;r<a.length;r++){switch(a.charCodeAt(r)){case 34:n="&quot;";break;case 38:n="&amp;";break;case 60:n="&lt;";break;case 62:n="&gt;";break;default:continue}c!==r&&(s+=a.substring(c,r)),c=r+1,s+=n}return c!==r?s+a.substring(c,r):s}
var pug_has_own_property=Object.prototype.hasOwnProperty;
var pug_match_html=/["&<>]/;
function pug_rethrow(e,n,r,t){if(!(e instanceof Error))throw e;if(!("undefined"==typeof window&&n||t))throw e.message+=" on line "+r,e;var o,a,i,s;try{t=t||require("fs").readFileSync(n,{encoding:"utf8"}),o=3,a=t.split("\n"),i=Math.max(r-o,0),s=Math.min(a.length,r+o)}catch(t){return e.message+=" - could not read from "+n+" ("+t.message+")",void pug_rethrow(e,null,r)}o=a.slice(i,s).map(function(e,n){var t=n+i+1;return(t==r?"  > ":"    ")+t+"| "+e}).join("\n"),e.path=n;try{e.message=(n||"Pug")+":"+r+"\n"+o+"\n\n"+e.message}catch(e){}throw e}function template(locals) {var pug_html = "", pug_mixins = {}, pug_interp;var pug_debug_filename, pug_debug_line;try {;
    var locals_for_with = (locals || {});
    
    (function (active, description, icon, id, title, variant) {
      ;pug_debug_line = 1;
pug_html = pug_html + "\u003Cbutton" + (pug_attr("class", pug_classes([{ 'mode-option': true, ['mode-option--' + variant]: true, 'is-active': active }], [true]), false, false)+" type=\"button\""+pug_attr("data-mode", id, true, false)) + "\u003E";
;pug_debug_line = 2;
pug_html = pug_html + "\u003Cspan class=\"mode-option__icon\"\u003E";
;pug_debug_line = 2;
pug_html = pug_html + (null == (pug_interp = icon) ? "" : pug_interp) + "\u003C\u002Fspan\u003E";
;pug_debug_line = 3;
pug_html = pug_html + "\u003Cspan class=\"mode-option__content\"\u003E";
;pug_debug_line = 4;
pug_html = pug_html + "\u003Cspan class=\"mode-option__title\"\u003E";
;pug_debug_line = 4;
pug_html = pug_html + (pug_escape(null == (pug_interp = title) ? "" : pug_interp)) + "\u003C\u002Fspan\u003E";
;pug_debug_line = 5;
if (description) {
;pug_debug_line = 6;
pug_html = pug_html + "\u003Cspan class=\"mode-option__description\"\u003E";
;pug_debug_line = 6;
pug_html = pug_html + (pug_escape(null == (pug_interp = description) ? "" : pug_interp)) + "\u003C\u002Fspan\u003E";
}
pug_html = pug_html + "\u003C\u002Fspan\u003E";
;pug_debug_line = 7;
pug_html = pug_html + "\u003Cspan class=\"mode-option__chevron\" aria-hidden=\"true\"\u003E\u003C\u002Fspan\u003E\u003C\u002Fbutton\u003E";
    }.call(this, "active" in locals_for_with ?
        locals_for_with.active :
        typeof active !== 'undefined' ? active : undefined, "description" in locals_for_with ?
        locals_for_with.description :
        typeof description !== 'undefined' ? description : undefined, "icon" in locals_for_with ?
        locals_for_with.icon :
        typeof icon !== 'undefined' ? icon : undefined, "id" in locals_for_with ?
        locals_for_with.id :
        typeof id !== 'undefined' ? id : undefined, "title" in locals_for_with ?
        locals_for_with.title :
        typeof title !== 'undefined' ? title : undefined, "variant" in locals_for_with ?
        locals_for_with.variant :
        typeof variant !== 'undefined' ? variant : undefined));
    ;} catch (err) {pug_rethrow(err, pug_debug_filename, pug_debug_line);};return pug_html;}
export default function render(locals){ return template(locals, runtime); }
