import runtime from 'pug-runtime';
function pug_attr(t,e,n,r){if(!1===e||null==e||!e&&("class"===t||"style"===t))return"";if(!0===e)return" "+(r?t:t+'="'+t+'"');var f=typeof e;return"object"!==f&&"function"!==f||"function"!=typeof e.toJSON||(e=e.toJSON()),"string"==typeof e||(e=JSON.stringify(e),n||-1===e.indexOf('"'))?(n&&(e=pug_escape(e))," "+t+'="'+e+'"'):" "+t+"='"+e.replace(/'/g,"&#39;")+"'"}
function pug_escape(e){var a=""+e,t=pug_match_html.exec(a);if(!t)return e;var r,c,n,s="";for(r=t.index,c=0;r<a.length;r++){switch(a.charCodeAt(r)){case 34:n="&quot;";break;case 38:n="&amp;";break;case 60:n="&lt;";break;case 62:n="&gt;";break;default:continue}c!==r&&(s+=a.substring(c,r)),c=r+1,s+=n}return c!==r?s+a.substring(c,r):s}
var pug_match_html=/["&<>]/;function template(locals) {var pug_html = "", pug_mixins = {}, pug_interp;;
    var locals_for_with = (locals || {});
    
    (function (description, id, requiresPath, title) {
      pug_html = pug_html + "\u003Cbutton" + (" class=\"option-card\""+" type=\"button\""+pug_attr("data-id", id, true, false)+pug_attr("data-requires", requiresPath, true, false)) + "\u003E\u003Cspan class=\"option-card__title\"\u003E" + (pug_escape(null == (pug_interp = title) ? "" : pug_interp)) + "\u003C\u002Fspan\u003E";
if (description) {
pug_html = pug_html + "\u003Cspan class=\"option-card__description\"\u003E" + (pug_escape(null == (pug_interp = description) ? "" : pug_interp)) + "\u003C\u002Fspan\u003E";
}
if (requiresPath) {
pug_html = pug_html + "\u003Cspan class=\"option-card__badge\"\u003EPath required\u003C\u002Fspan\u003E";
}
pug_html = pug_html + "\u003C\u002Fbutton\u003E";
    }.call(this, "description" in locals_for_with ?
        locals_for_with.description :
        typeof description !== 'undefined' ? description : undefined, "id" in locals_for_with ?
        locals_for_with.id :
        typeof id !== 'undefined' ? id : undefined, "requiresPath" in locals_for_with ?
        locals_for_with.requiresPath :
        typeof requiresPath !== 'undefined' ? requiresPath : undefined, "title" in locals_for_with ?
        locals_for_with.title :
        typeof title !== 'undefined' ? title : undefined));
    ;;return pug_html;}
export default function render(locals) { return template(locals, runtime); }
