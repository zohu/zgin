const char* build_time()
{
    static const char* psz_build_time = __DATE__ " " __TIME__ ;
    return psz_build_time;
}