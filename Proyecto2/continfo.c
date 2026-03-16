#include <linux/module.h>
#include <linux/kernel.h>
#include <linux/init.h>
#include <linux/proc_fs.h>
#include <linux/seq_file.h>
#include <linux/sched.h>
#include <linux/sched/signal.h>
#include <linux/mm.h>
#include <linux/sched/mm.h>

MODULE_LICENSE("GPL");
MODULE_AUTHOR("Daniel Salvatierra - 202202768");
MODULE_DESCRIPTION("Monitor de contenedores SO1 Proyecto 2");

#define PROC_NAME "continfo_pr2_so1_202202768"

static int continfo_show(struct seq_file *m, void *v)
{
    struct task_struct *task;
    struct mm_struct *mm;

    // Memoria Ram
    struct sysinfo si;
    si_meminfo(&si);

    unsigned long total_ram = si.totalram * si.mem_unit / 1024 / 1024;
    unsigned long free_ram = si.freeram * si.mem_unit / 1024 / 1024;
    unsigned long used_ram = total_ram - free_ram;

    //Json
    seq_printf(m, "{\n");
    seq_printf(m, "  \"total_ram_mb\": %lu,\n", total_ram);
    seq_printf(m, "  \"free_ram_mb\": %lu,\n", free_ram);
    seq_printf(m, "  \"used_ram_mb\": %lu,\n", used_ram);
    seq_printf(m, "  \"processes\": [\n");
    
    bool first = true;

    // Iterar sobre los procesos
    rcu_read_lock();
    for_each_process(task) {
        unsigned long vsz = 0, rss = 0;
        unsigned long mem_percent = 0;
        unsigned long long cpu_time = 0;

        mm = get_task_mm(task);
        if (mm) {
            vsz = mm->total_vm * PAGE_SIZE / 1024;           // KB
            rss = get_mm_rss(mm)  * PAGE_SIZE / 1024;        // KB

            if (si.totalram > 0)
                mem_percent = (get_mm_rss(mm) * 10000) / si.totalram;

            mmput(mm);
        }

        // CPU time
        cpu_time = task->utime + task->stime;

        if (!first) seq_printf(m, ",\n");
        first = false;

        seq_printf(m,
            "    {\"pid\": %d, \"name\": \"%s\", \"vsz\": %lu,"
            " \"rss\": %lu, \"mem_percent\": %lu, \"cpu_time\": %llu}",
            task->pid,
            task->comm,       // nombre del proceso (max 16 chars)
            vsz,
            rss,
            mem_percent,      // mem_percent / 100 = porcentaje real
            cpu_time
        );
    }

        rcu_read_unlock();

        seq_printf(m, "\n  ]\n}\n");
        return 0;
}

static int continfo_open(struct inode *inode, struct file *file) {
    return single_open(file, continfo_show, NULL);
}

// Estructura de operaciones para kernels
static const struct proc_ops continfo_ops = {
    .proc_open    = continfo_open,
    .proc_read    = seq_read,
    .proc_lseek   = seq_lseek,
    .proc_release = single_release,
};

static struct proc_dir_entry *proc_entry;

static int __init continfo_init(void) {
    proc_entry = proc_create(PROC_NAME, 0444, NULL, &continfo_ops);
    if (!proc_entry) {
        printk(KERN_ERR "continfo: Error al crear /proc/%s\n", PROC_NAME);
        return -ENOMEM;
    }
    printk(KERN_INFO "continfo: Módulo cargado. Leyendo en /proc/%s\n", PROC_NAME);
    return 0;
}

static void __exit continfo_exit(void) {
    proc_remove(proc_entry);
    printk(KERN_INFO "continfo: Módulo descargado.\n");
}

module_init(continfo_init);
module_exit(continfo_exit);